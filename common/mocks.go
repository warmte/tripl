package common

import "time"

func StringPtr(v string) *string {
	return &v
}

func Int64Ptr(v int64) *int64 {
	return &v
}

func TimePtr(v time.Time) *time.Time {
	return &v
}

func NewTestTicket() Ticket {
	return Ticket{
		ID:               101,
		UserID:           1,
		TripID:           Int64Ptr(201),
		Vehicle:          StringPtr("Самолёт"),
		DeparturePoint:   StringPtr("Москва"),
		ArrivalPoint:     StringPtr("Стамбул"),
		DepartureTime:    TimePtr(time.Date(2026, time.October, 15, 10, 30, 0, 0, time.UTC)),
		ArrivalTime:      TimePtr(time.Date(2026, time.October, 15, 14, 45, 0, 0, time.UTC)),
		Cost:             StringPtr("15000"),
		Currency:         StringPtr("RUB"),
		DepartureAddress: StringPtr("Аэропорт Шереметьево"),
		ArrivalAddress:   StringPtr("Istanbul Airport"),
		Seat:             StringPtr("12A"),
		Number:           StringPtr("SU2134"),
		Comment:          StringPtr("Место у окна"),
	}
}

func NewTestBooking() Booking {
	return Booking{
		ID:           102,
		UserID:       1,
		TripID:       Int64Ptr(201),
		City:         StringPtr("Стамбул"),
		Place:        StringPtr("Hotel Bosphorus"),
		CheckInDate:  TimePtr(time.Date(2026, time.October, 15, 0, 0, 0, 0, time.UTC)),
		CheckOutDate: TimePtr(time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC)),
		Cost:         StringPtr("30000"),
		Currency:     StringPtr("RUB"),
		Address:      StringPtr("Sultanahmet, Istanbul"),
		CheckInTime:  TimePtr(time.Date(2026, time.October, 15, 15, 0, 0, 0, time.UTC)),
		CheckOutTime: TimePtr(time.Date(2026, time.October, 18, 11, 0, 0, 0, time.UTC)),
		Number:       StringPtr("RES-2026-102"),
		Comment:      StringPtr("Завтрак включён"),
	}
}

func NewTestEvent() Event {
	return Event{
		ID:        103,
		UserID:    1,
		TripID:    Int64Ptr(201),
		Name:      StringPtr("Экскурсия по Босфору"),
		Location:  StringPtr("Порт"),
		StartTime: TimePtr(time.Date(2026, time.October, 16, 19, 0, 0, 0, time.UTC)),
		EndTime:   TimePtr(time.Date(2026, time.October, 16, 21, 0, 0, 0, time.UTC)),
		Cost:      StringPtr("2500"),
		Currency:  StringPtr("RUB"),
		Comment:   StringPtr("Билеты куплены заранее"),
	}
}

func NewTestTrip() Trip {
	return Trip{
		ID:        201,
		UserID:    1,
		Name:      "Поездка в Стамбул",
		StartDate: time.Date(2026, time.October, 15, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC),
		Comment:   StringPtr("Осенний отпуск"),
	}
}
