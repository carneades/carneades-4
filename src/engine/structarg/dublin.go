// Dublin Core Metadata
package engine

type Element int

const (
	Contributor Element = iota
	Coverage
	Creator
	Date
	Format
	Identifier
	Language // "en" "de", etc.
	Publisher
	Relation
	Rights
	Source
	Subject
	Title
	Type
	Key // non-standard element, for BibTeX keys
)

type Lang string

// Multiple values of elements are separated by semicolons
// Description elements are represented separately, to allow
// for translations in multiple langauges
type Metadata struct {
	Elements    map[Element]string
	Description map[Lang]string
}
