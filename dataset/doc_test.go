package dataset_test

import (
	"fmt"
	"github.com/clambin/simplejson/v3/dataset"
	"time"
)

func Example() {
	d := dataset.New()

	for day := 1; day < 5; day++ {
		d.Add(time.Date(2022, time.January, 5-day, 0, 0, 0, 0, time.UTC), "A", float64(5-day))
	}

	d.AddColumn("B", func(values map[string]float64) float64 {
		return values["A"] * 2
	})

	response := d.GenerateTableResponse()

	fmt.Printf("%v\n", response)
}
