package station

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Station struct {
	prices map[time.Time]uint64
}

func New(file string) (*Station, error) {

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read gas prices file: %w", err)
	}

	csvr := csv.NewReader(bytes.NewReader(data))
	records, err := csvr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not read gas price records: %w", err)
	}

	prices := make(map[time.Time]uint64, len(records))
	for _, record := range records[1:] {

		date, err := time.Parse("1/2/2006", record[0])
		if err != nil {
			return nil, fmt.Errorf("could not parse gas price date: %w", err)
		}

		value, err := strconv.ParseUint(record[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse gas price value: %w", err)
		}

		prices[date] = value
	}

	date := time.Date(2021, time.May, 9, 0, 0, 0, 0, time.UTC)
	_, ok := prices[date]
	if !ok {
		panic("wtf")
	}

	s := Station{
		prices: prices,
	}

	return &s, nil
}

func (s *Station) Gasprice(timestamp time.Time) (float64, error) {

	year, month, day := timestamp.Date()
	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	gasprice, ok := s.prices[date]
	if !ok {
		return 0, fmt.Errorf("gas price not found for date timestamp")
	}

	return float64(gasprice), nil
}
