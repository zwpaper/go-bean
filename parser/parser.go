package parser

// time checked payee description #tag ^link

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

type Configuration struct {
	Log      *log.Logger                           // Log is used to print warnings during parsing.
	ReadFile func(filename string) ([]byte, error) // ReadFile is used to read e.g. #+INCLUDE files.
}

type Options struct {
	Title             string
	OperatingCurrency []string
}

// Document contains the parsing results and a pointer to the Configuration.
type Document struct {
	*Configuration
	Path       string // Path of the file containing the parse input - used to resolve relative paths during parsing (e.g. INCLUDE).
	tokens     []token
	baseLvl    int
	Nodes      []Node
	NamedNodes map[string]Node
	Options    Options
	Error      error
}

// Node represents a parsed node of the document.
type Node interface {
	String() string // String returns the pretty printed Org mode string for the node (see OrgWriter).

	// node() make sure no others implement Node
	node()
}

type lexFn = func(line string) (t token, ok bool)
type parseFn = func(*Document, int, stopFn) (int, Node)
type stopFn = func(*Document, int) bool

type token struct {
	kind    string
	content string
	matches []string
}

var lexFns = []lexFn{
	lexComment,
	lexOption,
	lexAccount,
	lexPad,
	lexHeading,
	lexBody,
	lexBalance,
}

var nilToken = token{"nil", "", nil}

// New returns a new Configuration with (hopefully) sane defaults.
func New() *Configuration {
	return &Configuration{
		Log:      log.New(os.Stderr, "go-org: ", 0),
		ReadFile: ioutil.ReadFile,
	}
}

// Parse parses the input into an AST (and some other helpful fields like Outline).
// To allow method chaining, errors are stored in document.Error rather than being returned.
func (c *Configuration) Parse(input io.Reader, path string) (*Document, error) {
	d := &Document{
		Configuration: c,
		NamedNodes:    map[string]Node{},
		Path:          path,
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			logrus.Error("panic: ", recovered)
			d.Error = fmt.Errorf("could not parse input: %v", recovered)
		}
	}()
	if d.tokens != nil {
		d.Error = fmt.Errorf("parse was called multiple times")
	}
	if err := d.tokenize(input); err != nil {
		return nil, err
	}
	_, _, err := d.parseMany(0, func(d *Document, i int) bool { return i >= len(d.tokens) })
	if err != nil {
		return nil, err
	}

	return d, nil
}

// Silent disables all logging of warnings during parsing.
func (c *Configuration) Silent() *Configuration {
	c.Log = log.New(ioutil.Discard, "", 0)
	return c
}

func (d *Document) tokenize(input io.Reader) error {
	d.tokens = []token{}
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			// "skip empty line"
			continue
		}
		t, err := tokenize(line)
		if err != nil {
			return err
		}
		d.tokens = append(d.tokens, t)
	}
	if err := scanner.Err(); err != nil {
		d.Error = fmt.Errorf("could not tokenize input: %s", err)
		return d.Error
	}

	return nil
}

var plainTextRegexp = regexp.MustCompile(`^(\s*)(.*)`)

func (d *Document) parseOne(i int, stop stopFn) (consumed int, node Node, err error) {
	switch d.tokens[i].kind {
	case tokenComment:
		// Todo: skiped currently
		consumed++
	case tokenOption:
		consumed, node, err = d.parseOption(i, stop)
	case tokenAccount:
		consumed, node, err = d.parseAccount(i, stop)
	case tokenPad:
		consumed, node, err = d.parsePad(i, stop)
	case tokenHeading:
		consumed, node, err = d.parseHeading(i, stop)
	case tokenBody:
		consumed, node, err = d.parseBody(i, stop)
	case tokenBalance:
		consumed, node, err = d.parseBalance(i, stop)
	default:
		err = fmt.Errorf("not supported kind: %s", d.tokens[i].kind)
	}
	if err != nil || consumed == 0 {
		return i, nil, fmt.Errorf("can not parse %s: %w", d.tokens[i].kind, err)
	}

	return consumed, node, nil
}

func (d *Document) parseMany(i int, stop stopFn) (int, []Node, error) {
	start := i
	for i < len(d.tokens) && !stop(d, i) {
		consumed, node, err := d.parseOne(i, stop)
		if err != nil {
			return i, nil, err
		}
		i += consumed
		if node != nil {
			d.Nodes = append(d.Nodes, node)
		}
	}
	return i - start, d.Nodes, nil
}

func tokenize(line string) (token, error) {
	for _, lexFn := range lexFns {
		if token, ok := lexFn(line); ok {
			return token, nil
		}
	}
	return token{}, fmt.Errorf("could not lex line: %s", line)
}
