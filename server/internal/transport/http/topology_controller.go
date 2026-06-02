package http

import (
	"net/http"

	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"

	"github.com/fromforgesoftware/forge/server/internal/api"
	"github.com/fromforgesoftware/forge/server/internal/app"
)

// Permission actions for the platform topology surface.
const (
	permPlatformRead   = "platform.read"
	permPlatformManage = "platform.manage"
	permWorkloadDelete = "platform:workload.delete"
	permClusterManage  = "platform:cluster.manage"
)

// TopologyController serves the live cluster graph (read) and the audited
// maintenance actions (restart/scale/pause/resume/delete, node cordon/drain).
// The read surface is JSON:API via the kit handlers; the actions ride JSON:API
// command handlers and return the refreshed topology.
type TopologyController struct {
	topo  app.TopologyUsecase
	auth  app.AuthUsecase
	authz app.AuthzUsecase
}

func NewTopologyController(topo app.TopologyUsecase, auth app.AuthUsecase, authz app.AuthzUsecase) kitrest.Controller {
	return &TopologyController{topo: topo, auth: auth, authz: authz}
}

func (c *TopologyController) gate(action string, h http.Handler) http.Handler {
	return guard(c.auth, c.authz, action, h)
}

func (c *TopologyController) Routes(r kitrest.Router) {
	r.Method(http.MethodGet, "/api/platform/topology", c.gate(permPlatformRead,
		kitrest.NewJsonApiListHandler(c.topo, api.TopologyToDTO)))

	const wl = "/api/platform/workloads/{namespace}/{kind}/{name}"
	r.Method(http.MethodPost, wl+"/restart", c.gate(permPlatformManage,
		kitrest.NewJsonApiCommandHandler(c.topo.RestartWorkload, decodeWorkloadRef, api.TopologyToDTO,
			kitrest.HandlerWithSuccessStatus(http.StatusOK))))
	r.Method(http.MethodPost, wl+"/scale", c.gate(permPlatformManage,
		kitrest.NewJsonApiCommandHandler(c.topo.ScaleWorkload, decodeScale, api.TopologyToDTO,
			kitrest.HandlerWithSuccessStatus(http.StatusOK))))
	r.Method(http.MethodPost, wl+"/pause", c.gate(permPlatformManage,
		kitrest.NewJsonApiCommandHandler(c.topo.PauseWorkload, decodeWorkloadRef, api.TopologyToDTO,
			kitrest.HandlerWithSuccessStatus(http.StatusOK))))
	r.Method(http.MethodPost, wl+"/resume", c.gate(permPlatformManage,
		kitrest.NewJsonApiCommandHandler(c.topo.ResumeWorkload, decodeWorkloadRef, api.TopologyToDTO,
			kitrest.HandlerWithSuccessStatus(http.StatusOK))))
	r.Method(http.MethodDelete, wl, c.gate(permWorkloadDelete,
		kitrest.NewJsonApiCommandHandler(c.topo.DeleteWorkload, decodeWorkloadRef, api.TopologyToDTO,
			kitrest.HandlerWithSuccessStatus(http.StatusOK))))

	const nd = "/api/platform/nodes/{name}"
	r.Method(http.MethodPost, nd+"/cordon", c.gate(permClusterManage,
		kitrest.NewJsonApiCommandHandler(c.topo.CordonNode, decodeNodeCmd, api.TopologyToDTO,
			kitrest.HandlerWithSuccessStatus(http.StatusOK))))
	r.Method(http.MethodPost, nd+"/uncordon", c.gate(permClusterManage,
		kitrest.NewJsonApiCommandHandler(c.topo.UncordonNode, decodeNodeCmd, api.TopologyToDTO,
			kitrest.HandlerWithSuccessStatus(http.StatusOK))))
	r.Method(http.MethodPost, nd+"/drain", c.gate(permClusterManage,
		kitrest.NewJsonApiCommandHandler(c.topo.DrainNode, decodeNodeCmd, api.TopologyToDTO,
			kitrest.HandlerWithSuccessStatus(http.StatusOK))))
}

func decodeWorkloadRef(req *http.Request) (app.WorkloadRefCommand, error) {
	return app.WorkloadRefCommand{
		Kind:      req.PathValue("kind"),
		Namespace: req.PathValue("namespace"),
		Name:      req.PathValue("name"),
	}, nil
}

func decodeScale(req *http.Request) (app.WorkloadScaleCommand, error) {
	body, err := kitrest.UnmarshalPayloadFromRequest[*api.ScaleCommandDTO](req)
	if err != nil {
		return app.WorkloadScaleCommand{}, err
	}
	return app.WorkloadScaleCommand{
		Kind:      req.PathValue("kind"),
		Namespace: req.PathValue("namespace"),
		Name:      req.PathValue("name"),
		Replicas:  body.Replicas(),
	}, nil
}

func decodeNodeCmd(req *http.Request) (app.NodeCommand, error) {
	return app.NodeCommand{Name: req.PathValue("name")}, nil
}
