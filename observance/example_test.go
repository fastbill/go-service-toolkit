package observance

import "fmt"

func ExampleNewTestLogger() {
	logger := NewTestLogger()
	obs := &Obs{Logger: logger}
	obs.Logger.Error("testMessage1")
	fmt.Println(logger.LastEntry().Level, logger.LastEntry().Message)

	obs.Logger.Info("testMessage2")
	fmt.Printf("%+v", logger.Entries())

	// Output:
	// error testMessage1
	// [{Level:error Message:testMessage1} {Level:info Message:testMessage2}]
}
