package rodtemplate

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

type ElementSelector interface {
	El(selector string) *ElementTemplate
	Els(selector string) ElementsTemplate
	Has(selector string) bool
}

var _ ElementSelector = (*ElementTemplate)(nil)

type ElementTemplate struct {
	*rod.Element
}

func (e ElementTemplate) El(selector string) *ElementTemplate {
	return &ElementTemplate{Element: e.MustElement(selector)}
}

func (e ElementTemplate) Els(selector string) ElementsTemplate {
	return toElementsTemplate(e.MustElements(selector))
}

func (e ElementTemplate) Has(selector string) bool {
	has, _, err := e.Element.Has(selector)
	if err != nil {
		panic(err)
	}

	return has
}

// WaitUntilHasForYear
// calls WaitUntilHas(selector, time.Millisecond*10, time.Hour*24*365)
func (e ElementTemplate) WaitUntilHasForYear(selector string) bool {
	return e.WaitUntilHas(selector, time.Millisecond*10, time.Hour*24*365)
}

// WaitUntilHasForHour
// calls WaitUntilHas(selector, time.Millisecond*10, time.Hour)
func (e ElementTemplate) WaitUntilHasForHour(selector string) bool {
	return e.WaitUntilHas(selector, time.Millisecond*10, time.Hour)
}

func (e ElementTemplate) WaitUntilHas(selector string, sleepDuration time.Duration, deadline time.Duration) bool {
	deadlineTime := time.Now().Add(deadline)
	for time.Now().Before(deadlineTime) {
		if e.Has(selector) {
			return true
		}
		time.Sleep(sleepDuration)
	}

	return false
}

func toElementsTemplate(elements rod.Elements) ElementsTemplate {
	est := make([]*ElementTemplate, 0)
	for idx := range elements {
		est = append(est, &ElementTemplate{Element: elements[idx]})
	}

	return est
}

func NewElementsTemplate(elements rod.Elements) ElementsTemplate {
	return toElementsTemplate(elements)
}

func (e ElementTemplate) ElE(selector string) (*rod.Element, error) {
	return e.Element.Element(selector)
}

func (e ElementTemplate) ElementAttribute(selector string, name string) *string {
	return e.El(selector).MustAttribute(name)
}

func (e ElementTemplate) Height() float64 {
	quad := e.MustShape().Quads[0]

	return quad[7] - quad[1]
}

func (e ElementTemplate) SelectOrPanic(selector string) *ElementTemplate {
	if !e.Has(selector) {
		panic(fmt.Errorf("element is missing %s sub element", selector))
	}

	return e.El(selector)
}

func (e ElementTemplate) MustTextAsUInt64() uint64 {
	text := strings.TrimSpace(e.MustText())
	text = strings.ReplaceAll(text, ",", "")

	val, err := strconv.ParseUint(text, 0, 64)

	if err != nil {
		panic(err)
	}

	return val
}

func (e ElementTemplate) MustAttributeAsInt(attr string) int {
	attribute := e.MustAttribute(attr)
	if attribute == nil {
		panic(fmt.Errorf("attribute %s is not found in %s", attr, e.MustHTML()))
	}

	val, err := strconv.Atoi(strings.TrimSpace(*attribute))
	if err != nil {
		panic(err)
	}

	return val
}

func (e ElementTemplate) WaitEnabledAndWritable() error {
	err := e.WaitEnabled()
	if err != nil {
		return err
	}

	err = e.WaitWritable()
	if err != nil {
		return err
	}

	return nil
}

type ElementsTemplate []*ElementTemplate
