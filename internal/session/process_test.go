package session_test

import (
	"encoding/json"
	"time"

	dbMocks "github.com/satimoto/go-datastore/pkg/db/mocks"
	"github.com/satimoto/go-datastore/pkg/util"
	ferpMocks "github.com/satimoto/go-lnm/internal/ferp/mocks"
	"github.com/satimoto/go-lnm/internal/ito"
	lightningnetworkMocks "github.com/satimoto/go-lnm/internal/lightningnetwork/mocks"
	notificationMocks "github.com/satimoto/go-lnm/internal/notification/mocks"
	serviceMocks "github.com/satimoto/go-lnm/internal/service/mocks"
	sessionsMocks "github.com/satimoto/go-lnm/internal/session/mocks"
	ocpiMocks "github.com/satimoto/go-ocpi/pkg/ocpi/mocks"

	"testing"
)

func TestProcessChargingPeriods(t *testing.T) {
	cases := []struct {
		desc     string
		session  []byte
		tariff   []byte
		chargePower  float64
		location string
		date     string
		value    float64
	}{
		{
			desc: "No periods - time",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:39:09Z",
				"currency": "EUR",
				"total_cost": 0,
				"total_energy": 0.00,
				"charging_periods": [],
				"status": "ACTIVE",
				"last_updated": "2015-06-29T23:37:32Z"
			}`),
			tariff: []byte(`{
				"elements": [{
					"price_components": [{
						"type": "TIME",
						"price": 2.00,
						"step_size": 300
					}]
				}]
			}`),
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-29T23:37:32Z",
			value:    4,
		}, {
			desc: "No periods - time and energy",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:39:09Z",
				"currency": "EUR",
				"total_cost": 0,
				"total_energy": 0.00,
				"charging_periods": [],
				"status": "ACTIVE",
				"last_updated": "2015-06-29T23:37:32Z"
			}`),
			tariff: []byte(`{
				"elements": [{
					"price_components": [{
						"type": "TIME",
						"price": 2.00,
						"step_size": 300
					}]
				}, {
					"price_components": [{
						"type": "ENERGY",
						"price": 0.30,
						"step_size": 1
					}]
				}]
			}`),
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-29T23:37:32Z",
			value:    5.2,
		}, {
			desc: "No periods - kWh",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:39:00Z",
				"currency": "EUR",
				"total_cost": 0,
				"total_energy": 3.00,
				"charging_periods": [],
				"status": "ACTIVE",
				"last_updated": "2015-06-29T23:39:00Z"
			}`),
			tariff: []byte(`{
				"elements": [{
					"price_components": [{
						"type": "TIME",
						"price": 2.00,
						"step_size": 300
					}]
				}, {
					"price_components": [{
						"type": "ENERGY",
						"price": 0.30,
						"step_size": 1
					}]
				}]
			}`),
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-29T23:39:00Z",
			value:    4.9,
		}, {
			desc: "No periods - kWh + delta",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:39:00Z",
				"currency": "EUR",
				"total_cost": 0,
				"total_energy": 3.00,
				"charging_periods": [],
				"status": "ACTIVE",
				"last_updated": "2015-06-29T23:39:00Z"
			}`),
			tariff: []byte(`{
				"elements": [{
					"price_components": [{
						"type": "TIME",
						"price": 2.00,
						"step_size": 300
					}]
				}, {
					"price_components": [{
						"type": "ENERGY",
						"price": 0.30,
						"step_size": 1
					}]
				}]
			}`),
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-29T23:44:00Z",
			value:    5.067,
		}, {
			desc: "No periods - total cost",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:39:00Z",
				"currency": "EUR",
				"total_cost": 5,
				"total_energy": 3.00,
				"charging_periods": [],
				"status": "ACTIVE",
				"last_updated": "2015-06-29T23:39:00Z"
			}`),
			tariff: []byte(`{
				"elements": [{
					"price_components": [{
						"type": "TIME",
						"price": 2.00,
						"step_size": 300
					}]
				}, {
					"price_components": [{
						"type": "ENERGY",
						"price": 0.30,
						"step_size": 1
					}]
				}]
			}`),
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-29T23:37:32Z",
			value:    5,
		}, {
			desc: "No periods - total cost + delta",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:39:00Z",
				"currency": "EUR",
				"total_cost": 5,
				"total_energy": 3.00,
				"charging_periods": [],
				"status": "ACTIVE",
				"last_updated": "2015-06-29T23:39:00Z"
			}`),
			tariff: []byte(`{
				"elements": [{
					"price_components": [{
						"type": "TIME",
						"price": 2.00,
						"step_size": 300
					}]
				}, {
					"price_components": [{
						"type": "ENERGY",
						"price": 0.30,
						"step_size": 1
					}]
				}]
			}`),
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-29T23:44:00Z",
			value:    5.208333333333334,
		}, {
			desc: "Simple",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:39:09Z",
				"currency": "EUR",
				"total_cost": 4.00,
				"total_energy": 0.00,
				"charging_periods": [{
					"start_date_time": "2015-06-29T21:39:09Z",
					"dimensions": [{
						"type": "TIME",
						"volume": 1.973
					}]
				}],
				"status": "ACTIVE",
				"last_updated": "2015-06-29T23:37:32Z"
			}`),
			tariff: []byte(`{
				"elements": [{
					"price_components": [{
						"type": "TIME",
						"price": 2.00,
						"step_size": 300
					}]
				}]
			}`),
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-29T23:37:32Z",
			value:    4,
		}, {
			desc: "Simple",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:39:09Z",
				"currency": "EUR",
				"total_cost": 4.00,
				"total_energy": 0.00,
				"charging_periods": [{
					"start_date_time": "2015-06-29T21:39:09Z",
					"dimensions": [{
						"type": "TIME",
						"volume": 1.973
					}]
				}],
				"status": "ACTIVE",
				"last_updated": "2015-06-29T23:37:32Z"
			}`),
			tariff: []byte(`{
				"elements": [{
					"price_components": [{
						"type": "TIME",
						"price": 2.00,
						"step_size": 300
					}]
				}]
			}`),
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-29T23:37:32Z",
			value:    4,
		}, {
			desc: "Simple Session time",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:39:09Z",
				"currency": "EUR",
				"total_cost": 4.00,
				"total_energy": 0.00,
				"charging_periods": [{
					"start_date_time": "2015-06-29T21:39:09Z",
					"dimensions": [{
						"type": "TIME",
						"volume": 1.973
					}]
				}],
				"status": "ACTIVE",
				"last_updated": "2015-06-29T23:37:32Z"
			}`),
			tariff: []byte(`{
				"elements": [{
					"price_components": [{
						"type": "SESSION_TIME",
						"price": 2.00,
						"step_size": 300
					}]
				}]
			}`),
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-30T00:07:32Z",
			value:    5.013656201604957,
		}, {
			desc: "Complete - with total cost",
			session: []byte(`{
				"start_datetime": "2015-06-29T22:39:09Z",
				"end_datetime": "2015-06-29T23:50:16Z",
				"currency": "EUR",
				"total_cost": 8.50,
				"total_energy": 41.00,
				"charging_periods": [{
					"start_date_time": "2015-06-29T22:39:09Z",
					"dimensions": [{
						"type": "ENERGY",
						"volume": 12
					}, {
						"type": "MAX_CURRENT",
						"volume": 30
					}]
				}, {
					"start_date_time": "2015-06-29T22:40:54Z",
					"dimensions": [{
						"type": "ENERGY",
						"volume": 29
					}, {
						"type": "MIN_CURRENT",
						"volume": 34
					}]
				}, {
					"start_date_time": "2015-06-29T23:07:09Z",
					"dimensions": [{
						"type": "PARKING_TIME",
						"volume": 0.718
					}]
				}],
				"status": "COMPLETED",
				"last_updated": "2015-06-29T23:09:10Z"
			}`),
			tariff: []byte(`{
				"elements": [{
					"price_components": [{
						"type": "FLAT",
						"price": 2.50,
						"step_size": 1
					}]
				}, {
					"price_components": [{
						"type": "ENERGY",
						"price": 0.30,
						"step_size": 1
					}],
					"restrictions": {
						"max_power": 32.00
					}
				}, {
					"price_components": [{
						"type": "ENERGY",
						"price": 0.28,
						"step_size": 1
					}],
					"restrictions": {
						"min_power": 32.00,
						"day_of_week": ["MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY"]
					}
				}, {
					"price_components": [{
						"type": "ENERGY",
						"price": 0.26,
						"step_size": 1
					}],
					"restrictions": {
						"min_power": 32.00,
						"day_of_week": ["SATURDAY", "SUNDAY"]
					}
				}, {
					"price_components": [{
						"type": "PARKING_TIME",
						"price": 5.00,
						"step_size": 300
					}],
					"restrictions": {
						"start_time": "09:00",
						"end_time": "18:00",
						"day_of_week": ["MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY"]
					}
				}, {
					"price_components": [{
						"type": "PARKING_TIME",
						"price": 6.00,
						"step_size": 300
					}],
					"restrictions": {
						"start_time": "10:00",
						"end_time": "17:00",
						"day_of_week": ["SATURDAY"]
					}
				}]
			}`),
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-30T00:07:32Z",
			value:    8.5,
		}, {
			desc: "Complete - No total cost",
			session: []byte(`{
				"start_datetime": "2015-06-29T22:39:09Z",
				"end_datetime": "2015-06-29T23:50:16Z",
				"currency": "EUR",
				"total_cost": 0,
				"total_energy": 41.00,
				"charging_periods": [{
					"start_date_time": "2015-06-29T22:39:09Z",
					"dimensions": [{
						"type": "ENERGY",
						"volume": 12
					}, {
						"type": "MAX_CURRENT",
						"volume": 30
					}]
				}, {
					"start_date_time": "2015-06-29T22:40:54Z",
					"dimensions": [{
						"type": "ENERGY",
						"volume": 29
					}, {
						"type": "MIN_CURRENT",
						"volume": 34
					}]
				}, {
					"start_date_time": "2015-06-29T23:07:09Z",
					"dimensions": [{
						"type": "PARKING_TIME",
						"volume": 0.718
					}]
				}],
				"status": "COMPLETED",
				"last_updated": "2015-06-29T23:09:10Z"
			}`),
			tariff: []byte(`{
				"elements": [{
					"price_components": [{
						"type": "FLAT",
						"price": 2.50,
						"step_size": 1
					}]
				}, {
					"price_components": [{
						"type": "ENERGY",
						"price": 0.30,
						"step_size": 1
					}],
					"restrictions": {
						"max_power": 32.00
					}
				}, {
					"price_components": [{
						"type": "ENERGY",
						"price": 0.28,
						"step_size": 1
					}],
					"restrictions": {
						"min_power": 32.00,
						"day_of_week": ["MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY"]
					}
				}, {
					"price_components": [{
						"type": "ENERGY",
						"price": 0.26,
						"step_size": 1
					}],
					"restrictions": {
						"min_power": 32.00,
						"day_of_week": ["SATURDAY", "SUNDAY"]
					}
				}, {
					"price_components": [{
						"type": "PARKING_TIME",
						"price": 5.00,
						"step_size": 300
					}],
					"restrictions": {
						"start_time": "09:00",
						"end_time": "18:00",
						"day_of_week": ["MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY"]
					}
				}, {
					"price_components": [{
						"type": "PARKING_TIME",
						"price": 6.00,
						"step_size": 300
					}],
					"restrictions": {
						"start_time": "10:00",
						"end_time": "17:00",
						"day_of_week": ["SATURDAY"]
					}
				}]
			}`),
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-30T00:07:32Z",
			value:    14.22,
		}, {
			desc: "Complete - No total cost, incomplete charging periods",
			session: []byte(`{
				"start_datetime": "2015-06-29T22:39:09Z",
				"end_datetime": "2015-06-29T23:50:16Z",
				"currency": "EUR",
				"total_cost": 0,
				"total_energy": 41.00,
				"charging_periods": [{
					"start_date_time": "2015-06-29T23:07:09Z",
					"dimensions": [{
						"type": "PARKING_TIME",
						"volume": 0.718
					}]
				}],
				"status": "COMPLETED",
				"last_updated": "2015-06-29T23:09:10Z"
			}`),
			tariff: []byte(`{
				"elements": [{
					"price_components": [{
						"type": "FLAT",
						"price": 2.50,
						"step_size": 1
					}]
				}, {
					"price_components": [{
						"type": "ENERGY",
						"price": 0.30,
						"step_size": 1
					}],
					"restrictions": {
						"day_of_week": ["MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY"]
					}
				}, {
					"price_components": [{
						"type": "ENERGY",
						"price": 0.28,
						"step_size": 1
					}],
					"restrictions": {
						"day_of_week": ["SATURDAY", "SUNDAY"]
					}
				}, {
					"price_components": [{
						"type": "PARKING_TIME",
						"price": 5.00,
						"step_size": 300
					}],
					"restrictions": {
						"start_time": "09:00",
						"end_time": "18:00",
						"day_of_week": ["MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY"]
					}
				}, {
					"price_components": [{
						"type": "PARKING_TIME",
						"price": 6.00,
						"step_size": 300
					}],
					"restrictions": {
						"start_time": "10:00",
						"end_time": "17:00",
						"day_of_week": ["SATURDAY"]
					}
				}]
			}`),
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-30T00:07:32Z",
			value:    14.8,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			mockRepository := dbMocks.NewMockRepositoryService()
			mockFerpService := ferpMocks.NewService()
			mockLightningService := lightningnetworkMocks.NewService()
			mockNotificationService := notificationMocks.NewService()
			mockOcpiService := ocpiMocks.NewService()
			mockServices := serviceMocks.NewService(mockFerpService, mockLightningService, mockNotificationService, mockOcpiService)
			sessionResolver := sessionsMocks.NewResolver(mockRepository, mockServices)

			sessionIto := ito.SessionIto{}
			json.Unmarshal(tc.session, &sessionIto)

			tariffIto := ito.TariffIto{}
			json.Unmarshal(tc.tariff, &tariffIto)

			timeLocation, _ := time.LoadLocation(tc.location)

			processTime := util.ParseTime(tc.date, nil)
			value, _, _ := sessionResolver.ProcessChargingPeriods(&sessionIto, &tariffIto, tc.chargePower, timeLocation, *processTime)

			if value != tc.value {
				t.Errorf("Value mismatch: %v expecting %v", value, tc.value)
			}
		})
	}
}

func TestProcessChargingPeriods2(t *testing.T) {
	tariffBytes := []byte(`{
		"elements": [{
			"price_components": [{
				"type": "FLAT",
				"price": 2.50,
				"step_size": 1
			}]
		}, {
			"price_components": [{
				"type": "TIME",
				"price": 2.00,
				"step_size": 300
			}]
		}]
	}`)

	cases := []struct {
		desc     string
		session  []byte
		tariff   []byte
		chargePower  float64
		location string
		date     string
		value    float64
	}{
		{
			desc: "Simple",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:00:00Z",
				"kwh": 0.00,
				"currency": "EUR",
				"charging_periods": [],
				"total_cost": 0.00,
				"status": "ACTIVE",
				"last_updated": "2015-06-29T21:00:00Z"
			}`),
			tariff:   tariffBytes,
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-29T21:00:00Z",
			value:    2.5,
		}, {
			desc: "Simple",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:00:00Z",
				"kwh": 0.00,
				"currency": "EUR",
				"charging_periods": [{
					"start_date_time": "2015-06-29T21:01:00Z",
					"dimensions": [{
						"type": "TIME",
						"volume": 0.016
					}]
				}],
				"total_cost": 0.00,
				"status": "ACTIVE",
				"last_updated": "2015-06-29T21:02:00Z"
			}`),
			tariff:   tariffBytes,
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-29T21:02:00Z",
			value:    2.667,
		}, {
			desc: "Simple",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:00:00Z",
				"end_datetime": "0001-01-01 00:00:00",
				"kwh": 0.00,
				"currency": "EUR",
				"charging_periods": [{
					"start_date_time": "2015-06-29T21:01:00Z",
					"dimensions": [{
						"type": "TIME",
						"volume": 0.016
					}]
				}],
				"total_cost": 0.00,
				"status": "ACTIVE",
				"last_updated": "2015-06-29T21:02:00Z"
			}`),
			tariff:   tariffBytes,
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-29T21:02:00Z",
			value:    2.667,
		}, {
			desc: "Simple",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:00:00Z",
				"kwh": 0.00,
				"currency": "EUR",
				"charging_periods": [{
					"start_date_time": "2015-06-29T21:01:00Z",
					"dimensions": [{
						"type": "TIME",
						"volume": 0.016
					}]
				}],
				"total_cost": 0.00,
				"status": "ACTIVE",
				"last_updated": "2015-06-29T21:02:00Z"
			}`),
			tariff:   tariffBytes,
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-29T21:03:00Z",
			value:    2.667,
		}, {
			desc: "Simple",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:00:00Z",
				"kwh": 0.00,
				"currency": "EUR",
				"charging_periods": [{
					"start_date_time": "2015-06-29T21:01:00Z",
					"dimensions": [{
						"type": "TIME",
						"volume": 0.016
					}]
				}],
				"total_cost": 0.00,
				"status": "ACTIVE",
				"last_updated": "2015-06-29T21:02:00Z"
			}`),
			tariff:   tariffBytes,
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-29T21:07:00Z",
			value:    2.833,
		}, {
			desc: "Simple",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:00:00Z",
				"kwh": 0.00,
				"currency": "EUR",
				"charging_periods": [{
					"start_date_time": "2015-06-29T21:01:00Z",
					"dimensions": [{
						"type": "TIME",
						"volume": 0.016
					}]
				}],
				"total_cost": 0.00,
				"status": "ACTIVE",
				"last_updated": "2015-06-29T21:02:00Z"
			}`),
			tariff:   tariffBytes,
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-29T22:02:00Z",
			value:    4.5,
		}, {
			desc: "Simple",
			session: []byte(`{
				"start_datetime": "2015-06-29T21:00:00Z",
				"kwh": 0.00,
				"currency": "EUR",
				"charging_periods": [{
					"start_date_time": "2015-06-29T21:01:00Z",
					"dimensions": [{
						"type": "TIME",
						"volume": 1.0
					}]
				}],
				"total_cost": 0.00,
				"status": "ACTIVE",
				"last_updated": "2015-06-29T22:01:00Z"
			}`),
			tariff:   tariffBytes,
			chargePower:  11.040,
			location: "Europe/Berlin",
			date:     "2015-06-29T22:02:00Z",
			value:    4.667,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			mockRepository := dbMocks.NewMockRepositoryService()
			mockFerpService := ferpMocks.NewService()
			mockLightningService := lightningnetworkMocks.NewService()
			mockNotificationService := notificationMocks.NewService()
			mockOcpiService := ocpiMocks.NewService()
			mockServices := serviceMocks.NewService(mockFerpService, mockLightningService, mockNotificationService, mockOcpiService)
			sessionResolver := sessionsMocks.NewResolver(mockRepository, mockServices)

			sessionIto := ito.SessionIto{}
			json.Unmarshal(tc.session, &sessionIto)

			tariffIto := ito.TariffIto{}
			json.Unmarshal(tc.tariff, &tariffIto)

			timeLocation, _ := time.LoadLocation(tc.location)

			processTime := util.ParseTime(tc.date, nil)
			value, _, _ := sessionResolver.ProcessChargingPeriods(&sessionIto, &tariffIto, tc.chargePower, timeLocation, *processTime)

			if value != tc.value {
				t.Errorf("Value mismatch: %v expecting %v", value, tc.value)
			}
		})
	}
}
