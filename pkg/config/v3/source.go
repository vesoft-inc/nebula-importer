package configv3

import (
	"path/filepath"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/client"
	configbase "github.com/vesoft-inc/nebula-importer/v4/pkg/config/base"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/importer"
	specv3 "github.com/vesoft-inc/nebula-importer/v4/pkg/spec/v3"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/utils"
)

type (
	Source struct {
		configbase.Source `yaml:",inline"`
		Nodes             specv3.Nodes `yaml:"tags,omitempty"`
		Edges             specv3.Edges `yaml:"edges,omitempty"`
	}

	Sources []Source
)

func (s *Source) BuildGraph(graphName string, opts ...specv3.GraphOption) (*specv3.Graph, error) {
	options := make([]specv3.GraphOption, 0, len(s.Nodes)+len(s.Edges)+len(opts))
	for i := range s.Nodes {
		node := s.Nodes[i]
		options = append(options, specv3.WithGraphNodes(node))
	}
	for i := range s.Edges {
		edge := s.Edges[i]
		options = append(options, specv3.WithGraphEdges(edge))
	}
	options = append(options, opts...)
	graph := specv3.NewGraph(graphName, options...)
	graph.Complete()
	if err := graph.Validate(); err != nil {
		return nil, err
	}
	return graph, nil
}

func (s *Source) BuildImporters(graphName string, pool client.Pool) ([]importer.Importer, error) {
	graph, err := s.BuildGraph(graphName)
	if err != nil {
		return nil, err
	}
	importers := make([]importer.Importer, 0, len(s.Nodes)+len(s.Edges))
	for k := range s.Nodes {
		node := s.Nodes[k]
		builder := graph.NodeStatementBuilder(node)
		i := importer.New(builder, pool)
		importers = append(importers, i)
	}

	for k := range s.Edges {
		edge := s.Edges[k]
		builder := graph.EdgeStatementBuilder(edge)
		i := importer.New(builder, pool)
		importers = append(importers, i)
	}
	return importers, nil
}

// OptimizePath optimizes relative paths base to the configuration file path
func (ss Sources) OptimizePath(configPath string) error {
	configPathDir := filepath.Dir(configPath)
	for i := range ss {
		if ss[i].SourceConfig.Local != nil {
			ss[i].SourceConfig.Local.Path = utils.RelativePathBaseOn(configPathDir, ss[i].SourceConfig.Local.Path)
		}
	}
	return nil
}

// OptimizePathWildCard optimizes the wildcards in the paths
func (ss *Sources) OptimizePathWildCard() error {
	nss := make(Sources, 0, len(*ss))
	for i := range *ss {
		ssCpy := (*ss)[i]

		baseSources, isSupportGlob, err := (*ss)[i].Glob()
		if err != nil {
			return err
		}
		if isSupportGlob {
			for j := range baseSources {
				cpy := ssCpy
				cpy.Source = *baseSources[j]
				nss = append(nss, cpy)
			}
		} else {
			nss = append(nss, ssCpy)
		}
	}
	*ss = nss
	return nil
}
