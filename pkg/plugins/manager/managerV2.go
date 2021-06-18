package manager

import (
	"context"
	"fmt"

	pluginsV2 "github.com/grafana/grafana-plugin-manager/pkg/plugins"
	managerV2 "github.com/grafana/grafana-plugin-manager/pkg/plugins/manager"
	modelsV2 "github.com/grafana/grafana-plugin-manager/pkg/plugins/models"

	"github.com/grafana/grafana-plugin-sdk-go/backend"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/registry"
	"github.com/grafana/grafana/pkg/setting"
)

var _ pluginsV2.PluginManager = (*PluginManagerV2)(nil)

type PluginManagerV2 struct {
	Cfg                    *setting.Cfg                  `inject:""`
	License                models.Licensing              `inject:""`
	PluginRequestValidator models.PluginRequestValidator `inject:""`

	log     log.Logger
	manager *managerV2.PluginManager
}

func init() {
	registry.Register(&registry.Descriptor{
		Name:         "PluginManagerV2",
		Instance:     &PluginManagerV2{},
		InitPriority: registry.MediumHigh,
	})
}

func (m *PluginManagerV2) Init() error {
	m.log = log.New("plugin.managerv2")

	if m.IsDisabled() {
		m.log.Info("Plugin Manager V2 is disabled")
		return nil
	}

	m.log.Info("Plugin Manager V2 is starting...")
	innerManager := managerV2.New(
		m.Cfg.PluginsPath,
		m.Cfg.BuildVersion,
		m.Cfg.AppURL,
		m.Cfg.AppSubURL,
		m.log,
	)

	m.log.Info("Plugin Manager V2 has successfully started!")
	m.manager = innerManager

	return nil
}

func (m *PluginManagerV2) Start() error {
	return m.manager.Init()
}

func (m *PluginManagerV2) IsDisabled() bool {
	_, exists := m.Cfg.FeatureToggles["pluginManagerV2"]
	return !exists
}

func (m *PluginManagerV2) StartV2() error {
	if m.IsDisabled() {
		return fmt.Errorf("cannot start Plugin Manager V2 as the feature toggle is disabled")
	}

	return m.manager.Init()
}

func (m *PluginManagerV2) DataSource(pluginID string) {
	panic("implement me")
}

func (m *PluginManagerV2) Panel(pluginID string) {
	panic("implement me")
}

func (m *PluginManagerV2) App(pluginID string) {
	panic("implement me")
}

func (m *PluginManagerV2) Renderer() *pluginsV2.Plugin {
	panic("implement me")
}

func (m *PluginManagerV2) Plugins() {
	panic("implement me")
}

func (m *PluginManagerV2) DataSources() {
	panic("implement me")
}

func (m *PluginManagerV2) Panels() {
	panic("implement me")
}

func (m *PluginManagerV2) Apps() {
	panic("implement me")
}

func (m *PluginManagerV2) StaticRoutes() []*modelsV2.PluginStaticRoute {
	panic("implement me")
}

func (m *PluginManagerV2) Errors(pluginID string) {
	panic("implement me")
}

func (m *PluginManagerV2) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return m.manager.QueryData(ctx, req)
}

func (m *PluginManagerV2) CallResource(ctx context.Context, req *backend.CallResourceRequest) (*backend.CallResourceResponse, error) {
	panic("implement me")
}

func (m *PluginManagerV2) CollectMetrics(ctx context.Context, pluginID string) (*backend.CollectMetricsResult, error) {
	panic("implement me")
}

func (m *PluginManagerV2) CheckHealth(ctx context.Context, pCtx backend.PluginContext) (*backend.CheckHealthResult, error) {
	panic("implement me")
}

func (m *PluginManagerV2) IsRegistered(pluginID string) bool {
	return m.manager.IsRegistered(pluginID)
}

func (m *PluginManagerV2) Install(ctx context.Context, pluginID, version string) error {
	panic("implement me")
}

func (m *PluginManagerV2) Uninstall(ctx context.Context, pluginID string) error {
	panic("implement me")
}
