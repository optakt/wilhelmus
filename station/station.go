package station

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math/big"
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

	s := Station{
		prices: prices,
	}

	return &s, nil
}

func (s *Station) Gasprice(timestamp time.Time) (*big.Int, error) {

	gasPrice := big.NewInt(0)

	year, month, day := timestamp.Date()
	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	value, ok := s.prices[date]
	if !ok {
		return gasPrice, fmt.Errorf("gas price not found for date timestamp")
	}

	gasPrice.SetUint64(value)

	return gasPrice, nil
}
