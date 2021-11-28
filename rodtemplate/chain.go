package rodtemplate

import (
	"fmt"
	"strings"
)

func NewInspectChain(template *ElementTemplate) *InspectChain {
	return &InspectChain{
		et:        template,
		selectors: []string{},
		errors:    []error{},
	}
}

type InspectChain struct {
	skipNext  bool
	prev      *InspectChain
	et        *ElementTemplate
	selectors []string
	errors    []error

	returnSelf bool
}

type InspectOneFunc func(el *ElementTemplate) error
type InspectEachFunc func(idx int, el *ElementTemplate) error

//SelfChain ... returns self chain which returns self(chain) after ForOne
//which means no chain informations are preserved.
func (ic *InspectChain) SelfChain() *InspectChain {
	return &InspectChain{
		skipNext:  ic.skipNext,
		prev:      ic.prev,
		et:        ic.et,
		selectors: ic.selectors,
		errors:    ic.errors,

		returnSelf: true,
	}

}

//ForOne ... find one element, call f and return found element chain
func (ic *InspectChain) ForOne(selector string, panicIfNot, stopOnError bool, f InspectOneFunc) *InspectChain {
	selectors := append(ic.selectors, selector)

	if ic.skipNext {
		return &InspectChain{
			skipNext:  true,
			prev:      ic,
			et:        nil,
			selectors: selectors,
			errors:    append(ic.errors, nil),
		}
	}

	var err error
	var el *ElementTemplate

	if ic.et == nil {
		err = fmt.Errorf("element is nil for %s", strings.Join(ic.selectors, " > "))
	} else if ic.et.Has(selector) {
		el = ic.et.El(selector)
		err = f(el)
	} else if panicIfNot {
		panic(fmt.Errorf("%s is not found", strings.Join(selectors, " > ")))
	} else {
		err = f(nil)
	}

	if ic.returnSelf {
		return &InspectChain{
			skipNext:  ic.skipNext,
			prev:      ic.prev,
			et:        ic.et,
			selectors: ic.selectors,
			errors:    ic.errors,

			returnSelf: true,
		}
	}

	return &InspectChain{
		skipNext:  err != nil && stopOnError,
		prev:      ic,
		et:        el,
		selectors: selectors,
		errors:    append(ic.errors, err),
	}
}

//ForEach ... find elements, call f for each elements and return self(chain)
func (ic *InspectChain) ForEach(selector string, panicIfNot, stopOnError bool, f InspectEachFunc) *InspectChain {
	selectors := append(ic.selectors, selector)

	if ic.skipNext {
		return &InspectChain{
			skipNext:  true,
			prev:      ic.prev,
			et:        nil,
			selectors: selectors,
			errors:    append(ic.errors, nil),
		}
	}

	var err error
	var els ElementsTemplate

	if ic.et == nil {
		err = fmt.Errorf("element is nil for %s", strings.Join(ic.selectors, " > "))
	} else if ic.et.Has(selector) {
		els = ic.et.Els(selector)
		for idx, el := range els {
			err = f(idx, el)
		}
	} else if panicIfNot {
		panic(fmt.Errorf("%s is not found", strings.Join(selectors, " > ")))
	} else {
		//TODO: how to repeat?
	}

	return &InspectChain{
		skipNext:  err != nil && stopOnError,
		prev:      ic.prev,
		et:        ic.et,
		selectors: selectors,
		errors:    append(ic.errors, err),
	}
}
