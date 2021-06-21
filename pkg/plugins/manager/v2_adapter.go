package manager

import (
	pluginsV2 "github.com/grafana/grafana-plugin-manager/pkg/plugins"
	"github.com/grafana/grafana/pkg/plugins"
)

func fromV2(v2 pluginsV2.Plugin) *plugins.PluginBase {
	return &plugins.PluginBase{
		Type: v2.Type,
		Name: v2.Name,
		Id:   v2.ID,
		Info: plugins.PluginInfo{
			Author: plugins.PluginInfoLink{
				Name: v2.Info.Author.Name,
				Url:  v2.Info.Author.Url,
			},
			Description: v2.Info.Description,
			Links:       v2.Info.Links,
			Logos: plugins.PluginLogos{
				Small: v2.Info.Logos.Small,
				Large: v2.Info.Logos.Large,
			},
			Build: plugins.PluginBuildInfo{
				Time:   v2.Info.Build.Time,
				Repo:   v2.Info.Build.Repo,
				Branch: v2.Info.Build.Branch,
				Hash:   v2.Info.Build.Hash,
			},
			Screenshots: v2.Info.Screenshots,
			Version:     v2.Info.Version,
			Updated:     v2.Info.Updated,
		},
		Dependencies: plugins.PluginDependencies{
			GrafanaVersion: v2.Dependencies.GrafanaVersion,
			Plugins: []plugins.PluginDependencyItem{
				{
					Type:    v2.Dependencies.Plugins[0].Type,
					Id:      v2.Dependencies.Plugins[0].Id,
					Name:    v2.Dependencies.Plugins[0].Name,
					Version: v2.Dependencies.Plugins[0].Version,
				},
			},
		},
		Includes: []*plugins.PluginInclude{
			{
				Name:       v2.Includes[0].Name,
				Path:       v2.Includes[0].Path,
				Type:       v2.Includes[0].Type,
				Component:  v2.Includes[0].Component,
				Role:       v2.Includes[0].Role,
				AddToNav:   v2.Includes[0].AddToNav,
				DefaultNav: v2.Includes[0].DefaultNav,
				Slug:       v2.Includes[0].Slug,
				Icon:       v2.Includes[0].Icon,
				Id:         v2.Includes[0].Id,
			},
		},
		Module:              v2.Module,
		BaseUrl:             v2.BaseUrl,
		Category:            v2.Category,
		HideFromList:        v2.HideFromList,
		Preload:             v2.Preload,
		State:               v2.State,
		Signature:           v2.Signature,
		Backend:             v2.Backend,
		IncludedInAppId:     v2.IncludedInAppID,
		PluginDir:           v2.PluginDir,
		DefaultNavUrl:       v2.DefaultNavURL,
		IsCorePlugin:        v2.IsCorePlugin(),
		SignatureType:       v2.SignatureType,
		SignatureOrg:        v2.SignatureOrg,
		GrafanaNetVersion:   v2.GrafanaComVersion,
		GrafanaNetHasUpdate: v2.GrafanaComHasUpdate,
		Root:                fromV2(v2.Parent),
	}
}
