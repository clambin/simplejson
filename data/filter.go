package data

import (
	"github.com/clambin/simplejson/v3/query"
	"time"
)

// Filter returns a Dataset meeting the provided query Args. Currently, it filters based on the args' time Range.
// only the first time column is taken into consideration.
func (t Table) Filter(args query.Args) (filtered *Table) {
	index, found := t.getFirstTimestampColumn()
	if !found {
		panic("unable to determine timestamp column")
	}
	f, _ := t.Frame.FilterRowsByField(index, func(i interface{}) (bool, error) {
		if !args.Args.Range.From.IsZero() && i.(time.Time).Before(args.Args.Range.From) {
			return false, nil
		}
		if !args.Args.Range.To.IsZero() && i.(time.Time).After(args.Args.Range.To) {
			return false, nil
		}
		return true, nil
	})
	return &Table{Frame: f}
}
