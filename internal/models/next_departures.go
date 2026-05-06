package models

import "time"

type NextDepartureResponse struct {
	Results     []NextDeparture `json:"results"`
	Attribution []string        `json:"attribution"`
}

type NextDeparture struct {
	LineName            string     `json:"line_name"`
	Destination         string     `json:"destination"`
	OperatorRef         string     `json:"operator_ref,omitempty"`
	AimedArrivalTime    *time.Time `json:"aimed_arrival_time,omitempty"`
	ExpectedArrivalTime *time.Time `json:"expected_arrival_time,omitempty"`
}
