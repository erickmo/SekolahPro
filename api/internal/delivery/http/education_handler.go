package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	createcourse "github.com/yourorg/boilerplate/internal/command/create_education_course"
	updatecourse "github.com/yourorg/boilerplate/internal/command/update_education_course"
	deletecourse "github.com/yourorg/boilerplate/internal/command/delete_education_course"
	createenrollment "github.com/yourorg/boilerplate/internal/command/create_enrollment"
	completeenrollment "github.com/yourorg/boilerplate/internal/command/complete_enrollment"
	recordkpi "github.com/yourorg/boilerplate/internal/command/record_education_kpi"
	getcourse "github.com/yourorg/boilerplate/internal/query/get_education_course_by_id"
	listcourses "github.com/yourorg/boilerplate/internal/query/list_education_courses"
	getenrollment "github.com/yourorg/boilerplate/internal/query/get_enrollment_by_id"
	listenrollments "github.com/yourorg/boilerplate/internal/query/list_enrollments"
	listkpis "github.com/yourorg/boilerplate/internal/query/list_education_kpis"
	"github.com/yourorg/boilerplate/internal/domain/education"
	"github.com/yourorg/boilerplate/internal/domain/enrollment"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/querybus"
)

// EducationHandler menangani HTTP request untuk domain Education (courses + enrollments + kpis).
type EducationHandler struct {
	BaseHandler
}

var courseSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"title":       "title",
		"created_at":  "created_at",
		"category":    "category",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

var enrollmentSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"enrolled_at": "enrolled_at",
		"created_at":  "created_at",
		"status":      "status",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

var kpiSortConfig = SortConfig{
	AllowedFields: map[string]string{
		"period_month": "period_month",
		"created_at":   "created_at",
		"metric_type":  "metric_type",
	},
	DefaultField: "created_at",
	DefaultOrder: "desc",
}

// NewEducationHandler membuat instance EducationHandler baru.
func NewEducationHandler(cb *commandbus.CommandBus, qb *querybus.QueryBus) *EducationHandler {
	return &EducationHandler{BaseHandler: NewBaseHandler(cb, qb)}
}

// RegisterRoutes mendaftarkan semua route untuk domain Education.
func (h *EducationHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/education/courses", func(r chi.Router) {
		r.Get("/", h.ListCourses)
		r.Post("/", h.CreateCourse)
		r.Get("/{id}", h.GetCourseByID)
		r.Put("/{id}", h.UpdateCourse)
		r.Delete("/{id}", h.DeleteCourse)
	})
	r.Route("/api/v1/education/enrollments", func(r chi.Router) {
		r.Get("/", h.ListEnrollments)
		r.Post("/", h.CreateEnrollment)
		r.Get("/{id}", h.GetEnrollmentByID)
		r.Put("/{id}/complete", h.CompleteEnrollment)
	})
	r.Route("/api/v1/education/kpis", func(r chi.Router) {
		r.Get("/", h.ListKPIs)
		r.Post("/", h.RecordKPI)
	})
}

// ── Course endpoints ──────────────────────────────────────────────────────────

// ListCourses godoc
// @Summary      List education courses
// @Tags         education-courses
// @Produce      json
// @Success      200  {object}  listcourses.Result
// @Router       /api/v1/education/courses [get]
func (h *EducationHandler) ListCourses(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, courseSortConfig)
	result, err := querybus.Dispatch[*listcourses.Result](r.Context(), h.queryBus, listcourses.Query{
		Params: params,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data kursus")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetCourseByID godoc
// @Summary      Get education course by ID
// @Tags         education-courses
// @Produce      json
// @Param        id   path      string  true  "Course ID"
// @Success      200  {object}  getcourse.Result
// @Router       /api/v1/education/courses/{id} [get]
func (h *EducationHandler) GetCourseByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getcourse.Result](r.Context(), h.queryBus, getcourse.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// CreateCourse godoc
// @Summary      Create education course
// @Tags         education-courses
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/education/courses [post]
func (h *EducationHandler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title         string   `json:"title"`
		Description   string   `json:"description"`
		CourseType    string   `json:"course_type"`
		Category      string   `json:"category"`
		DurationHours int      `json:"duration_hours"`
		ContentURL    string   `json:"content_url"`
		PassingScore  float64  `json:"passing_score"`
		MaxAttempts   int      `json:"max_attempts"`
		MandatoryFor  []string `json:"mandatory_for"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	mandatoryFor := req.MandatoryFor
	if mandatoryFor == nil {
		mandatoryFor = []string{}
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createcourse.Command{
		Title:         req.Title,
		Description:   req.Description,
		CourseType:    education.CourseType(req.CourseType),
		Category:      req.Category,
		DurationHours: req.DurationHours,
		ContentURL:    req.ContentURL,
		PassingScore:  req.PassingScore,
		MaxAttempts:   req.MaxAttempts,
		MandatoryFor:  mandatoryFor,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// UpdateCourse godoc
// @Summary      Update education course
// @Tags         education-courses
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Course ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/education/courses/{id} [put]
func (h *EducationHandler) UpdateCourse(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		Title         string   `json:"title"`
		Description   string   `json:"description"`
		CourseType    string   `json:"course_type"`
		Category      string   `json:"category"`
		DurationHours int      `json:"duration_hours"`
		ContentURL    string   `json:"content_url"`
		IsActive      bool     `json:"is_active"`
		PassingScore  float64  `json:"passing_score"`
		MaxAttempts   int      `json:"max_attempts"`
		MandatoryFor  []string `json:"mandatory_for"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	mandatoryFor := req.MandatoryFor
	if mandatoryFor == nil {
		mandatoryFor = []string{}
	}

	if err := h.commandBus.Dispatch(r.Context(), updatecourse.Command{
		ID: id,
		Title:         req.Title,
		Description:   req.Description,
		CourseType:    education.CourseType(req.CourseType),
		Category:      req.Category,
		DurationHours: req.DurationHours,
		ContentURL:    req.ContentURL,
		IsActive:      req.IsActive,
		PassingScore:  req.PassingScore,
		MaxAttempts:   req.MaxAttempts,
		MandatoryFor:  mandatoryFor,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// DeleteCourse godoc
// @Summary      Delete education course
// @Tags         education-courses
// @Param        id   path      string  true  "Course ID"
// @Success      204
// @Router       /api/v1/education/courses/{id} [delete]
func (h *EducationHandler) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), deletecourse.Command{ID: id}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Enrollment endpoints ──────────────────────────────────────────────────────

// ListEnrollments godoc
// @Summary      List education enrollments
// @Tags         education-enrollments
// @Produce      json
// @Success      200  {object}  listenrollments.Result
// @Router       /api/v1/education/enrollments [get]
func (h *EducationHandler) ListEnrollments(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, enrollmentSortConfig)
	result, err := querybus.Dispatch[*listenrollments.Result](r.Context(), h.queryBus, listenrollments.Query{
		Params: params,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data pendaftaran")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// GetEnrollmentByID godoc
// @Summary      Get enrollment by ID
// @Tags         education-enrollments
// @Produce      json
// @Param        id   path      string  true  "Enrollment ID"
// @Success      200  {object}  getenrollment.Result
// @Router       /api/v1/education/enrollments/{id} [get]
func (h *EducationHandler) GetEnrollmentByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	result, err := querybus.Dispatch[*getenrollment.Result](r.Context(), h.queryBus, getenrollment.Query{ID: id})
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// CreateEnrollment godoc
// @Summary      Create enrollment
// @Tags         education-enrollments
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/education/enrollments [post]
func (h *EducationHandler) CreateEnrollment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CourseID  string `json:"course_id"`
		NasabahID string `json:"nasabah_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	courseID, err := uuid.Parse(req.CourseID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "course_id tidak valid")
		return
	}
	nasabahID, err := uuid.Parse(req.NasabahID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "nasabah_id tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, createenrollment.Command{
		CourseID:  courseID,
		NasabahID: nasabahID,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}

// CompleteEnrollment godoc
// @Summary      Complete enrollment
// @Tags         education-enrollments
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Enrollment ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/education/enrollments/{id}/complete [put]
func (h *EducationHandler) CompleteEnrollment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var req struct {
		Score          float64 `json:"score"`
		CertificateURL string  `json:"certificate_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}
	if err := h.commandBus.Dispatch(r.Context(), completeenrollment.Command{
		ID:             id,
		Score:          req.Score,
		CertificateURL: req.CertificateURL,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]string{"id": id.String()})
}

// ── KPI endpoints ─────────────────────────────────────────────────────────────

// ListKPIs godoc
// @Summary      List education KPIs
// @Tags         education-kpis
// @Produce      json
// @Success      200  {object}  listkpis.Result
// @Router       /api/v1/education/kpis [get]
func (h *EducationHandler) ListKPIs(w http.ResponseWriter, r *http.Request) {
	params := h.ParsePagination(r, kpiSortConfig)
	result, err := querybus.Dispatch[*listkpis.Result](r.Context(), h.queryBus, listkpis.Query{
		Params: params,
	})
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "gagal mengambil data KPI")
		return
	}
	h.RespondJSON(w, http.StatusOK, result)
}

// RecordKPI godoc
// @Summary      Record education KPI
// @Tags         education-kpis
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]string
// @Router       /api/v1/education/kpis [post]
func (h *EducationHandler) RecordKPI(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MetricType  string  `json:"metric_type"`
		PeriodMonth string  `json:"period_month"`
		Value       float64 `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	ctx, holder := commandbus.WithResultID(r.Context())
	if err := h.commandBus.Dispatch(ctx, recordkpi.Command{
		MetricType:  enrollment.MetricType(req.MetricType),
		PeriodMonth: req.PeriodMonth,
		Value:       req.Value,
	}); err != nil {
		h.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	h.RespondJSON(w, http.StatusCreated, map[string]string{"id": holder.ID.String()})
}
