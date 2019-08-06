package observance

import "fmt"

func ExampleNewTestLogger() {
	logger := NewTestLogger()
	obs := &Obs{Logger: logger}
	obs.Logger.WithField("testField", "testValue").Error("testMessage1")
	fmt.Println(logger.LastEntry().Level, logger.LastEntry().Message, logger.LastEntry().Data)

	obs.Logger.Info("testMessage2")
	fmt.Printf("%+v", logger.Entries())

	// Output:
	// error testMessage1 map[testField:testValue]
	// [{Level:error Message:testMessage1 Data:map[testField:testValue]} {Level:info Message:testMessage2 Data:map[]}]
}
