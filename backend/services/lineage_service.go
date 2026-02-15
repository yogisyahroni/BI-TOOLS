package services

import (
	"fmt"
	"insight-engine-backend/database"
	"insight-engine-backend/models"
	"regexp"
	"sync"

	"gorm.io/gorm"
)

type LineageNode struct {
	ID    string `json:"id"`
	Type  string `json:"type"` // "source", "table", "query", "dashboard"
	Label string `json:"label"`
}

type LineageEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type LineageGraph struct {
	Nodes []LineageNode `json:"nodes"`
	Edges []LineageEdge `json:"edges"`
}

type LineageService struct {
	db *gorm.DB
}

var (
	lineageInstance *LineageService
	lineageOnce     sync.Once
)

func GetLineageService() *LineageService {
	lineageOnce.Do(func() {
		lineageInstance = &LineageService{
			db: database.DB,
		}
	})
	return lineageInstance
}

// GetLineageGraph builds the full lineage graph
func (s *LineageService) GetLineageGraph() (*LineageGraph, error) {
	graph := &LineageGraph{
		Nodes: []LineageNode{},
		Edges: []LineageEdge{},
	}

	// 1. Fetch DataSources
	var connections []models.Connection
	if err := s.db.Find(&connections).Error; err != nil {
		return nil, err
	}

	for _, conn := range connections {
		nodeId := fmt.Sprintf("ds-%s", conn.ID)
		graph.Nodes = append(graph.Nodes, LineageNode{
			ID:    nodeId,
			Type:  "source",
			Label: conn.Name,
		})
	}

	// 2. Fetch Queries
	var queries []models.SavedQuery
	if err := s.db.Find(&queries).Error; err != nil {
		return nil, err
	}

	tableRegex := regexp.MustCompile(`(?i)(?:FROM|JOIN)\s+([a-zA-Z0-9_."]+)`)

	for _, q := range queries {
		queryNodeId := fmt.Sprintf("q-%s", q.ID)
		graph.Nodes = append(graph.Nodes, LineageNode{
			ID:    queryNodeId,
			Type:  "query",
			Label: q.Name,
		})

		dsNodeId := fmt.Sprintf("ds-%s", q.ConnectionID)

		// Extract Tables
		matches := tableRegex.FindAllStringSubmatch(q.SQL, -1)
		uniqueTables := make(map[string]bool)

		for _, match := range matches {
			if len(match) > 1 {
				tableName := match[1]
				uniqueTables[tableName] = true
			}
		}

		if len(uniqueTables) == 0 {
			// If no table found, link DS -> Query directly
			// Check if DS node exists first (handling potential orphans)
			dsExists := false
			for _, n := range graph.Nodes {
				if n.ID == dsNodeId {
					dsExists = true
					break
				}
			}

			if dsExists {
				graph.Edges = append(graph.Edges, LineageEdge{
					ID:     fmt.Sprintf("e-%s-%s", dsNodeId, queryNodeId),
					Source: dsNodeId,
					Target: queryNodeId,
				})
			}
		} else {
			for tableName := range uniqueTables {
				tableNodeId := fmt.Sprintf("tbl-%s-%s", q.ConnectionID, tableName)

				// Add Table Node if not exists
				exists := false
				for _, n := range graph.Nodes {
					if n.ID == tableNodeId {
						exists = true
						break
					}
				}
				if !exists {
					graph.Nodes = append(graph.Nodes, LineageNode{
						ID:    tableNodeId,
						Type:  "table",
						Label: tableName,
					})

					// Edge: DS -> Table
					graph.Edges = append(graph.Edges, LineageEdge{
						ID:     fmt.Sprintf("e-%s-%s", dsNodeId, tableNodeId),
						Source: dsNodeId,
						Target: tableNodeId,
					})
				}

				// Edge: Table -> Query
				graph.Edges = append(graph.Edges, LineageEdge{
					ID:     fmt.Sprintf("e-%s-%s", tableNodeId, queryNodeId),
					Source: tableNodeId,
					Target: queryNodeId,
				})
			}
		}
	}

	// 3. Fetch Dashboards
	var dashboards []models.Dashboard
	if err := s.db.Preload("Cards").Find(&dashboards).Error; err != nil {
		return nil, err
	}

	for _, d := range dashboards {
		dashNodeId := fmt.Sprintf("d-%s", d.ID)
		graph.Nodes = append(graph.Nodes, LineageNode{
			ID:    dashNodeId,
			Type:  "dashboard",
			Label: d.Name,
		})

		// Edge: Query -> Dashboard
		seenQueryEdges := make(map[string]bool)
		for _, card := range d.Cards {
			if card.QueryID != nil {
				queryNodeId := fmt.Sprintf("q-%s", *card.QueryID)
				edgeKey := fmt.Sprintf("%s-%s", queryNodeId, dashNodeId)

				if !seenQueryEdges[edgeKey] {
					// Check if Query node exists
					queryExists := false
					for _, n := range graph.Nodes {
						if n.ID == queryNodeId {
							queryExists = true
							break
						}
					}

					if queryExists {
						graph.Edges = append(graph.Edges, LineageEdge{
							ID:     fmt.Sprintf("e-%s", edgeKey),
							Source: queryNodeId,
							Target: dashNodeId,
						})
						seenQueryEdges[edgeKey] = true
					}
				}
			}
		}
	}

	return graph, nil
}
