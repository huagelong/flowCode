package http

import (
	"context"
	"net/http"

	planningapp "anserflow/internal/app/planning"

	"github.com/gin-gonic/gin"
)

type PlanTaskHandler struct {
	app   planningapp.Service
	scope ScopeChecker
}

func NewPlanTaskHandler(app planningapp.Service, scope ScopeChecker) *PlanTaskHandler {
	return &PlanTaskHandler{app: app, scope: scope}
}

func (h *PlanTaskHandler) List(c *gin.Context) {
	orgID, ok := mustUint64Param(c, "org_id")
	if !ok {
		return
	}
	planID, ok := mustUint64Param(c, "plan_id")
	if !ok {
		return
	}
	if err := h.ensurePlanScope(c.Request.Context(), orgID, planID); err != nil {
		respondError(c, http.StatusForbidden, err)
		return
	}
	out, err := h.app.ListTasks(c.Request.Context(), planID)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, out)
}

func (h *PlanTaskHandler) Get(c *gin.Context) {
	orgID, ok := mustUint64Param(c, "org_id")
	if !ok {
		return
	}
	planID, ok := mustUint64Param(c, "plan_id")
	if !ok {
		return
	}
	taskID, ok := mustUint64Param(c, "task_id")
	if !ok {
		return
	}
	if err := h.ensurePlanScope(c.Request.Context(), orgID, planID); err != nil {
		respondError(c, http.StatusForbidden, err)
		return
	}
	out, err := h.app.GetTask(c.Request.Context(), planID, taskID)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, out)
}

func (h *PlanTaskHandler) Compile(c *gin.Context) {
	orgID, ok := mustUint64Param(c, "org_id")
	if !ok {
		return
	}
	planID, ok := mustUint64Param(c, "plan_id")
	if !ok {
		return
	}
	taskID, ok := mustUint64Param(c, "task_id")
	if !ok {
		return
	}
	if err := h.ensurePlanScope(c.Request.Context(), orgID, planID); err != nil {
		respondError(c, http.StatusForbidden, err)
		return
	}
	out, err := h.app.CompileTask(c.Request.Context(), planningapp.CompileTaskInput{
		PlanID:     planID,
		PlanTaskID: taskID,
	})
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, out)
}

type createIssueReq struct {
	UserID       uint64 `json:"user_id"`
	AutoDispatch bool   `json:"auto_dispatch"`
}

func (h *PlanTaskHandler) CreateIssue(c *gin.Context) {
	orgID, ok := mustUint64Param(c, "org_id")
	if !ok {
		return
	}
	planID, ok := mustUint64Param(c, "plan_id")
	if !ok {
		return
	}
	taskID, ok := mustUint64Param(c, "task_id")
	if !ok {
		return
	}
	if err := h.ensurePlanScope(c.Request.Context(), orgID, planID); err != nil {
		respondError(c, http.StatusForbidden, err)
		return
	}

	var req createIssueReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	issue, spec, err := h.app.CreateIssue(c.Request.Context(), planningapp.CreateIssueInput{
		PlanID:       planID,
		PlanTaskID:   taskID,
		CreatedBy:    req.UserID,
		AutoDispatch: req.AutoDispatch,
	})
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	respondCreated(c, gin.H{
		"issue": issue,
		"spec":  spec,
	})
}

func (h *PlanTaskHandler) ensurePlanScope(ctx context.Context, orgID uint64, planID uint64) error {
	ok, err := h.scope.PlanInOrg(ctx, orgID, planID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}
