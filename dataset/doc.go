/*
Package dataset makes it easier to produce time-based responses when dealing with data that may not necessarily be sequential.

A dataset holds a table of rows for each timestamp that is added to the dataset. When adding new columns, empty cells are
automatically added to the table for existing rows:

	d := dataset.New()          // creates an empty dataset
	d.Add(time.Now(), "A", 1)   // dataset has one row with a single cell, set to 1
    d.Add(time.Now(), "B", 2)   // dataset now has two rows. First row is 1, 0. Second row is 0, 2

Furthermore, dataset allows to add a new column, calculated from the values of the other columns:

	d := dataset.New()
	// add rows with values for columns "A" and "B"

	d.AddColumn("C", func(values map[string]float64) float64 {
		return values["A"] + values["B"]
	})
	// dataset now has a column "C", with the sum of columns "A" and "B"
*/
package dataset
