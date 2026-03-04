package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/domain/flight"
	"github.com/go-chi/chi/v5"
)

type FlightController interface {
	Create(http.ResponseWriter, *http.Request)
	GetByID(http.ResponseWriter, *http.Request)
	List(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
}

type flightController struct {
	module flight.Module
}

func NewFlightController(module flight.Module) *flightController {
	return &flightController{module: module}
}

func (c *flightController) Create(w http.ResponseWriter, r *http.Request) {
	req := createFlightRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	departureAt, err := parseDateTime(req.DepartureAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid departure_at format, expected RFC3339")
		return
	}
	arrivalAt, err := parseDateTime(req.ArrivalAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid arrival_at format, expected RFC3339")
		return
	}

	input := flight.CreateFlightInput{
		Origin:      req.Origin,
		Destination: req.Destination,
		DepartureAt: departureAt,
		ArrivalAt:   arrivalAt,
		TotalSeats:  req.TotalSeats,
		PriceCents:  req.PriceCents,
		Currency:    req.Currency,
		Status:      req.Status,
	}

	created, err := c.module.Create(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (c *flightController) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	item, err := c.module.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, flight.ErrFlightNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get flight")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (c *flightController) List(w http.ResponseWriter, r *http.Request) {
	limit := parseIntQuery(r, "limit", 20)
	offset := parseIntQuery(r, "offset", 0)

	items, err := c.module.List(r.Context(), flight.ListParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list flights")
		return
	}

	writeJSON(w, http.StatusOK, listFlightsResponse{Items: items})
}

func (c *flightController) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := updateFlightRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := flight.UpdateFlightInput{
		Origin:         req.Origin,
		Destination:    req.Destination,
		TotalSeats:     req.TotalSeats,
		AvailableSeats: req.AvailableSeats,
		PriceCents:     req.PriceCents,
		Currency:       req.Currency,
		Status:         req.Status,
	}
	if req.DepartureAt != nil {
		t, err := parseDateTime(*req.DepartureAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid departure_at format, expected RFC3339")
			return
		}
		input.DepartureAt = &t
	}
	if req.ArrivalAt != nil {
		t, err := parseDateTime(*req.ArrivalAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid arrival_at format, expected RFC3339")
			return
		}
		input.ArrivalAt = &t
	}

	updated, err := c.module.Update(r.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, flight.ErrFlightNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (c *flightController) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := c.module.Delete(r.Context(), id); err != nil {
		if errors.Is(err, flight.ErrFlightNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete flight")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Message: message})
}

func parseIntQuery(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}
