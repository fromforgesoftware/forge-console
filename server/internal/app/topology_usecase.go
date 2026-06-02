package app

import (
	"context"

	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/application/usecase"
	"github.com/fromforgesoftware/go-kit/filter"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"
)

// actorCtxKey carries the authenticated subject id into the usecase so cluster
// actions can be audited with who performed them.
type actorCtxKey struct{}

// WithActor stamps the acting subject id onto the context (set by the transport
// guard once the request is authenticated).
func WithActor(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, actorCtxKey{}, id)
}

// ActorFromContext returns the acting subject id, or "unknown" when unset.
func ActorFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(actorCtxKey{}).(string); ok && id != "" {
		return id
	}
	return "unknown"
}

// ActionAuditor records a mutating cluster action. A logging implementation is
// wired by default; a hallmark-backed sink can replace it without touching the
// usecase.
type ActionAuditor interface {
	RecordTopologyAction(ctx context.Context, actor, verb, target string)
}

// TopologyUsecase is the read surface for the graph plus the audited cluster
// maintenance actions. Each command performs the kubernetes mutation, audits
// it, then returns the refreshed topology so the UI reflects the new state.
type TopologyUsecase interface {
	repository.Getter[Topology]
	repository.Lister[Topology]

	RestartWorkload(ctx context.Context, cmd WorkloadRefCommand) (Topology, error)
	ScaleWorkload(ctx context.Context, cmd WorkloadScaleCommand) (Topology, error)
	PauseWorkload(ctx context.Context, cmd WorkloadRefCommand) (Topology, error)
	ResumeWorkload(ctx context.Context, cmd WorkloadRefCommand) (Topology, error)
	DeleteWorkload(ctx context.Context, cmd WorkloadRefCommand) (Topology, error)
	CordonNode(ctx context.Context, cmd NodeCommand) (Topology, error)
	UncordonNode(ctx context.Context, cmd NodeCommand) (Topology, error)
	DrainNode(ctx context.Context, cmd NodeCommand) (Topology, error)
}

type topologyUsecase struct {
	repository.Getter[Topology]
	repository.Lister[Topology]

	repo    TopologyRepository
	actions ClusterActions
	auditor ActionAuditor
}

func NewTopologyUsecase(repo TopologyRepository, actions ClusterActions, auditor ActionAuditor) TopologyUsecase {
	return &topologyUsecase{
		Getter:  usecase.NewGetter[Topology](repo, ResourceTypeTopology),
		Lister:  usecase.NewLister[Topology](repo),
		repo:    repo,
		actions: actions,
		auditor: auditor,
	}
}

func (uc *topologyUsecase) current(ctx context.Context) (Topology, error) {
	return uc.repo.Get(ctx, search.WithQueryOpts(query.FilterBy(filter.OpEq, "id", TopologyID)))
}

func (uc *topologyUsecase) audit(ctx context.Context, verb, target string) {
	if uc.auditor != nil {
		uc.auditor.RecordTopologyAction(ctx, ActorFromContext(ctx), verb, target)
	}
}

func (uc *topologyUsecase) RestartWorkload(ctx context.Context, cmd WorkloadRefCommand) (Topology, error) {
	if err := uc.actions.RestartWorkload(ctx, cmd); err != nil {
		return nil, err
	}
	uc.audit(ctx, "restart", cmd.Namespace+"/"+cmd.Kind+"/"+cmd.Name)
	return uc.current(ctx)
}

func (uc *topologyUsecase) ScaleWorkload(ctx context.Context, cmd WorkloadScaleCommand) (Topology, error) {
	if err := uc.actions.ScaleWorkload(ctx, cmd); err != nil {
		return nil, err
	}
	uc.audit(ctx, "scale", cmd.Namespace+"/"+cmd.Kind+"/"+cmd.Name)
	return uc.current(ctx)
}

func (uc *topologyUsecase) PauseWorkload(ctx context.Context, cmd WorkloadRefCommand) (Topology, error) {
	if err := uc.actions.PauseWorkload(ctx, cmd); err != nil {
		return nil, err
	}
	uc.audit(ctx, "pause", cmd.Namespace+"/"+cmd.Kind+"/"+cmd.Name)
	return uc.current(ctx)
}

func (uc *topologyUsecase) ResumeWorkload(ctx context.Context, cmd WorkloadRefCommand) (Topology, error) {
	if err := uc.actions.ResumeWorkload(ctx, cmd); err != nil {
		return nil, err
	}
	uc.audit(ctx, "resume", cmd.Namespace+"/"+cmd.Kind+"/"+cmd.Name)
	return uc.current(ctx)
}

func (uc *topologyUsecase) DeleteWorkload(ctx context.Context, cmd WorkloadRefCommand) (Topology, error) {
	if err := uc.actions.DeleteWorkload(ctx, cmd); err != nil {
		return nil, err
	}
	uc.audit(ctx, "delete", cmd.Namespace+"/"+cmd.Kind+"/"+cmd.Name)
	return uc.current(ctx)
}

func (uc *topologyUsecase) CordonNode(ctx context.Context, cmd NodeCommand) (Topology, error) {
	if err := uc.actions.CordonNode(ctx, cmd); err != nil {
		return nil, err
	}
	uc.audit(ctx, "cordon", "node/"+cmd.Name)
	return uc.current(ctx)
}

func (uc *topologyUsecase) UncordonNode(ctx context.Context, cmd NodeCommand) (Topology, error) {
	if err := uc.actions.UncordonNode(ctx, cmd); err != nil {
		return nil, err
	}
	uc.audit(ctx, "uncordon", "node/"+cmd.Name)
	return uc.current(ctx)
}

func (uc *topologyUsecase) DrainNode(ctx context.Context, cmd NodeCommand) (Topology, error) {
	if err := uc.actions.DrainNode(ctx, cmd); err != nil {
		return nil, err
	}
	uc.audit(ctx, "drain", "node/"+cmd.Name)
	return uc.current(ctx)
}
