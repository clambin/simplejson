package data

import (
	"github.com/clambin/simplejson/v6"
	"time"
)

// Filter returns a Dataset meeting the provided query QueryArgs. Currently, it filters based on the args' time Range.
// only the first time column is taken into consideration.
func (t Table) Filter(args simplejson.Args) (filtered *Table) {
	index, found := t.getFirstTimestampColumn()
	if !found {
		return &Table{Frame: t.Frame.EmptyCopy()}
	}

	f, _ := t.Frame.FilterRowsByField(index, func(i interface{}) (bool, error) {
		if !args.Range.From.IsZero() && i.(time.Time).Before(args.Range.From) {
			return false, nil
		}
		if !args.Range.To.IsZero() && i.(time.Time).After(args.Range.To) {
			return false, nil
		}
		return true, nil
	})

	return &Table{Frame: f}
}
