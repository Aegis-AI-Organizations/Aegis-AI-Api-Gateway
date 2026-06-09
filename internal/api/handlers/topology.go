package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
	"github.com/gin-gonic/gin"
)

const (
	defaultNeo4jURL      = "http://aegis-neo4j-mvp.aegis-system.svc.cluster.local:7474"
	defaultNeo4jUser     = "neo4j"
	defaultNeo4jPassword = "neo4j_password"
	defaultNeo4jDatabase = "neo4j"
	publicAPIHostname    = "api.aegis-ai.fr"
)

type TopologyResponse struct {
	Hosts []TopologyHost `json:"hosts"`
}

type TopologyHost struct {
	ID          string              `json:"id"`
	Hostname    string              `json:"hostname,omitempty"`
	IPAddresses []string            `json:"ip_addresses,omitempty"`
	Containers  []TopologyContainer `json:"containers,omitempty"`
	Processes   []TopologyProcess   `json:"processes,omitempty"`
}

type TopologyContainer struct {
	ID           string            `json:"id"`
	Name         string            `json:"name,omitempty"`
	Image        string            `json:"image,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	Ports        []TopologyPort    `json:"ports,omitempty"`
	ExposedPorts []TopologyPort    `json:"exposed_ports,omitempty"`
	Processes    []TopologyProcess `json:"processes,omitempty"`
}

type TopologyProcess struct {
	PID         int32   `json:"pid"`
	Name        string  `json:"name,omitempty"`
	CommandLine *string `json:"command_line,omitempty"`
	User        *string `json:"user,omitempty"`
}

type TopologyPort struct {
	Number   int32   `json:"number"`
	Protocol string  `json:"protocol"`
	State    *string `json:"state,omitempty"`
}

type Neo4jTopologyService struct {
	url      string
	user     string
	password string
	database string
	client   *http.Client
}

type topologyHostBuilder struct {
	host       TopologyHost
	containers map[string]*topologyContainerBuilder
	processes  map[string]TopologyProcess
}

type topologyContainerBuilder struct {
	container TopologyContainer
	processes map[string]TopologyProcess
}

type neo4jTxError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type neo4jTxRow struct {
	Row []any `json:"row"`
}

type neo4jTxResult struct {
	Data []neo4jTxRow `json:"data"`
}

type neo4jTxResponse struct {
	Results []neo4jTxResult `json:"results"`
	Errors  []neo4jTxError  `json:"errors"`
}

type neo4jTxStatement struct {
	Statement  string         `json:"statement"`
	Parameters map[string]any `json:"parameters"`
}

type neo4jTxRequest struct {
	Statements []neo4jTxStatement `json:"statements"`
}

func NewNeo4jTopologyServiceFromEnv() TopologyService {
	return NewNeo4jTopologyService(
		envOrDefault("NEO4J_URL", defaultNeo4jURL),
		envOrDefault("NEO4J_USER", defaultNeo4jUser),
		envOrDefault("NEO4J_PASSWORD", defaultNeo4jPassword),
		envOrDefault("NEO4J_DATABASE", defaultNeo4jDatabase),
		nil,
	)
}

func NewNeo4jTopologyService(url, user, password, database string, client *http.Client) *Neo4jTopologyService {
	if client == nil {
		client = &http.Client{
			Timeout: 15 * time.Second,
		}
	}
	return &Neo4jTopologyService{
		url:      strings.TrimSpace(url),
		user:     strings.TrimSpace(user),
		password: strings.TrimSpace(password),
		database: strings.TrimSpace(database),
		client:   client,
	}
}

func (a *API) GetTopologyHandler(c *gin.Context) {
	if a.Topology == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Topology service unavailable"})
		return
	}

	companyID := c.Param("id")
	if companyID == "" {
		if value, ok := c.Get(string(types.CompanyIDKey)); ok {
			companyID = fmt.Sprint(value)
		}
	}
	if companyID == "" {
		companyID = c.Query("company_id")
	}

	companyID = strings.TrimSpace(companyID)
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}

	resp, err := a.Topology.GetTopology(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load topology"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Neo4jTopologyService) GetTopology(ctx context.Context, companyID string) (TopologyResponse, error) {
	companyID = strings.TrimSpace(companyID)
	if companyID == "" {
		return TopologyResponse{}, fmt.Errorf("company_id is required")
	}

	hosts := map[string]*topologyHostBuilder{}

	if err := s.loadHosts(ctx, companyID, hosts); err != nil {
		return TopologyResponse{}, err
	}
	if err := s.loadContainers(ctx, companyID, hosts); err != nil {
		return TopologyResponse{}, err
	}
	if err := s.loadHostProcesses(ctx, companyID, hosts); err != nil {
		return TopologyResponse{}, err
	}
	if err := s.loadContainerProcesses(ctx, companyID, hosts); err != nil {
		return TopologyResponse{}, err
	}

	s.ensurePublicApiHost(hosts)

	return TopologyResponse{Hosts: renderHosts(hosts)}, nil
}

func (s *Neo4jTopologyService) loadHosts(
	ctx context.Context,
	companyID string,
	hosts map[string]*topologyHostBuilder,
) error {
	rows, err := s.executeQuery(ctx, `
		MATCH (h:Host)
		WHERE h.companyId = $company_id
		RETURN
		  h.id AS id,
		  coalesce(h.hostname, h.rawId, h.id) AS hostname,
		  coalesce(h.ipAddresses, []) AS ip_addresses
		ORDER BY hostname ASC
	`, map[string]any{"company_id": companyID})
	if err != nil {
		return err
	}

	for _, row := range rows {
		if len(row) < 3 {
			continue
		}
		hostID := stringValue(row[0])
		if hostID == "" {
			continue
		}
		host := ensureHostBuilder(hosts, hostID)
		host.host.ID = hostID
		host.host.Hostname = stringValue(row[1])
		host.host.IPAddresses = uniqueStrings(asStringSlice(row[2]))
	}

	return nil
}

func (s *Neo4jTopologyService) loadContainers(
	ctx context.Context,
	companyID string,
	hosts map[string]*topologyHostBuilder,
) error {
	rows, err := s.executeQuery(ctx, `
		MATCH (h:Host)-[:RUNS_CONTAINER]->(c:Container)
		WHERE h.companyId = $company_id
		RETURN
		  h.id AS host_id,
		  coalesce(h.hostname, h.rawId, h.id) AS host_hostname,
		  coalesce(h.ipAddresses, []) AS host_ip_addresses,
		  c.id AS container_id,
		  coalesce(c.name, c.rawId, c.id) AS container_name,
		  coalesce(c.image, "") AS image,
		  coalesce(c.env, []) AS env,
		  coalesce(c.ports, []) AS ports,
		  coalesce(c.exposedPorts, []) AS exposed_ports
		ORDER BY host_hostname ASC, container_name ASC
	`, map[string]any{"company_id": companyID})
	if err != nil {
		return err
	}

	for _, row := range rows {
		if len(row) < 9 {
			continue
		}

		hostID := stringValue(row[0])
		containerID := stringValue(row[3])
		if hostID == "" || containerID == "" {
			continue
		}

		host := ensureHostBuilder(hosts, hostID)
		if host.host.Hostname == "" {
			host.host.Hostname = stringValue(row[1])
		}
		if len(host.host.IPAddresses) == 0 {
			host.host.IPAddresses = uniqueStrings(asStringSlice(row[2]))
		}

		container := host.containers[containerID]
		if container == nil {
			container = &topologyContainerBuilder{
				container: TopologyContainer{ID: containerID},
				processes: map[string]TopologyProcess{},
			}
			host.containers[containerID] = container
		}

		container.container.Name = firstNonEmpty(
			container.container.Name,
			stringValue(row[4]),
		)
		container.container.Image = firstNonEmpty(container.container.Image, stringValue(row[5]))
		if len(container.container.Env) == 0 {
			container.container.Env = parseEnv(row[6])
		}
		if len(container.container.Ports) == 0 {
			container.container.Ports = parsePorts(row[7])
		}
		if len(container.container.ExposedPorts) == 0 {
			container.container.ExposedPorts = parsePorts(row[8])
		}
	}

	return nil
}

func (s *Neo4jTopologyService) loadHostProcesses(
	ctx context.Context,
	companyID string,
	hosts map[string]*topologyHostBuilder,
) error {
	rows, err := s.executeQuery(ctx, `
		MATCH (h:Host)-[:RUNS_PROCESS]->(p:Process)
		WHERE h.companyId = $company_id
		RETURN
		  h.id AS host_id,
		  p.id AS process_id,
		  p.pid AS pid,
		  p.name AS name,
		  p.commandLine AS command_line,
		  p.user AS user
		ORDER BY p.pid ASC, p.name ASC
	`, map[string]any{"company_id": companyID})
	if err != nil {
		return err
	}

	for _, row := range rows {
		if len(row) < 6 {
			continue
		}
		hostID := stringValue(row[0])
		if hostID == "" {
			continue
		}
		host := ensureHostBuilder(hosts, hostID)
		process := processFromRow(row[2], row[3], row[4], row[5])
		key := stringValue(row[1])
		if key == "" {
			key = processKey(process)
		}
		host.processes[key] = process
	}

	return nil
}

func (s *Neo4jTopologyService) loadContainerProcesses(
	ctx context.Context,
	companyID string,
	hosts map[string]*topologyHostBuilder,
) error {
	rows, err := s.executeQuery(ctx, `
		MATCH (h:Host)-[:RUNS_CONTAINER]->(c:Container)-[:RUNS_PROCESS]->(p:Process)
		WHERE h.companyId = $company_id
		RETURN
		  h.id AS host_id,
		  c.id AS container_id,
		  p.id AS process_id,
		  p.pid AS pid,
		  p.name AS name,
		  p.commandLine AS command_line,
		  p.user AS user
		ORDER BY p.pid ASC, p.name ASC
	`, map[string]any{"company_id": companyID})
	if err != nil {
		return err
	}

	for _, row := range rows {
		if len(row) < 7 {
			continue
		}
		hostID := stringValue(row[0])
		containerID := stringValue(row[1])
		if hostID == "" || containerID == "" {
			continue
		}
		host := ensureHostBuilder(hosts, hostID)
		container := host.containers[containerID]
		if container == nil {
			container = &topologyContainerBuilder{
				container: TopologyContainer{ID: containerID},
				processes: map[string]TopologyProcess{},
			}
			host.containers[containerID] = container
		}
		process := processFromRow(row[3], row[4], row[5], row[6])
		key := stringValue(row[2])
		if key == "" {
			key = processKey(process)
		}
		container.processes[key] = process
	}

	return nil
}

func (s *Neo4jTopologyService) ensurePublicApiHost(hosts map[string]*topologyHostBuilder) {
	for _, host := range hosts {
		if strings.EqualFold(host.host.Hostname, publicAPIHostname) || host.host.ID == "api-route" {
			return
		}
	}

	hosts["api-route"] = &topologyHostBuilder{
		host: TopologyHost{
			ID:       "api-route",
			Hostname: publicAPIHostname,
		},
		containers: map[string]*topologyContainerBuilder{},
		processes:  map[string]TopologyProcess{},
	}
}

func renderHosts(hosts map[string]*topologyHostBuilder) []TopologyHost {
	rendered := make([]TopologyHost, 0, len(hosts))
	for _, host := range hosts {
		rendered = append(rendered, renderHost(host))
	}

	sort.SliceStable(rendered, func(i, j int) bool {
		if rendered[i].Hostname == rendered[j].Hostname {
			return rendered[i].ID < rendered[j].ID
		}
		return rendered[i].Hostname < rendered[j].Hostname
	})

	return rendered
}

func renderHost(builder *topologyHostBuilder) TopologyHost {
	host := builder.host

	processes := make([]TopologyProcess, 0, len(builder.processes))
	for _, process := range builder.processes {
		processes = append(processes, process)
	}
	sort.SliceStable(processes, func(i, j int) bool {
		if processes[i].PID == processes[j].PID {
			return processes[i].Name < processes[j].Name
		}
		return processes[i].PID < processes[j].PID
	})
	host.Processes = processes

	containers := make([]TopologyContainer, 0, len(builder.containers))
	for _, container := range builder.containers {
		containers = append(containers, renderContainer(container))
	}
	sort.SliceStable(containers, func(i, j int) bool {
		if containers[i].Name == containers[j].Name {
			return containers[i].ID < containers[j].ID
		}
		return containers[i].Name < containers[j].Name
	})
	host.Containers = containers

	return host
}

func renderContainer(builder *topologyContainerBuilder) TopologyContainer {
	container := builder.container

	processes := make([]TopologyProcess, 0, len(builder.processes))
	for _, process := range builder.processes {
		processes = append(processes, process)
	}
	sort.SliceStable(processes, func(i, j int) bool {
		if processes[i].PID == processes[j].PID {
			return processes[i].Name < processes[j].Name
		}
		return processes[i].PID < processes[j].PID
	})
	container.Processes = processes

	return container
}

func ensureHostBuilder(hosts map[string]*topologyHostBuilder, hostID string) *topologyHostBuilder {
	if builder, ok := hosts[hostID]; ok {
		if builder.containers == nil {
			builder.containers = map[string]*topologyContainerBuilder{}
		}
		if builder.processes == nil {
			builder.processes = map[string]TopologyProcess{}
		}
		if builder.host.ID == "" {
			builder.host.ID = hostID
		}
		return builder
	}

	builder := &topologyHostBuilder{
		host:       TopologyHost{ID: hostID},
		containers: map[string]*topologyContainerBuilder{},
		processes:  map[string]TopologyProcess{},
	}
	hosts[hostID] = builder
	return builder
}

func processFromRow(pidValue any, nameValue any, commandValue any, userValue any) TopologyProcess {
	process := TopologyProcess{
		PID:  int32Value(pidValue),
		Name: stringValue(nameValue),
	}
	if command := stringValue(commandValue); command != "" {
		process.CommandLine = &command
	}
	if user := stringValue(userValue); user != "" {
		process.User = &user
	}
	return process
}

func processKey(process TopologyProcess) string {
	return strings.ToLower(
		strings.Join(
			[]string{
				strconv.FormatInt(int64(process.PID), 10),
				process.Name,
				derefString(process.CommandLine),
				derefString(process.User),
			},
			"|",
		),
	)
}

func parseEnv(value any) map[string]string {
	if value == nil {
		return map[string]string{}
	}
	if envMap, ok := value.(map[string]any); ok {
		result := map[string]string{}
		for key, raw := range envMap {
			trimmedKey := strings.TrimSpace(key)
			if trimmedKey == "" {
				continue
			}
			result[trimmedKey] = stringValue(raw)
		}
		return result
	}

	items := asStringSlice(value)
	result := map[string]string{}
	for _, item := range items {
		if !strings.Contains(item, "=") {
			continue
		}
		key, val, _ := strings.Cut(item, "=")
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		result[key] = val
	}
	return result
}

func parsePorts(value any) []TopologyPort {
	items := asInterfaceSlice(value)
	if len(items) == 0 {
		return nil
	}

	ports := make([]TopologyPort, 0, len(items))
	seen := map[int32]struct{}{}
	for _, item := range items {
		var portNumber int32
		var protocol = "tcp"
		var state *string

		switch raw := item.(type) {
		case map[string]any:
			portNumber = int32Value(
				firstAny(
					raw["number"],
					raw["container_port"],
					raw["containerPort"],
					raw["host_port"],
					raw["hostPort"],
					raw["port"],
				),
			)
			protocol = strings.ToLower(stringValue(firstAny(raw["protocol"], "tcp")))
			if rawState := stringValue(raw["state"]); rawState != "" {
				state = &rawState
			}
		case string:
			parts := strings.Split(raw, ":")
			if len(parts) > 0 {
				portNumber = int32Value(parts[0])
			}
			if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
				protocol = strings.ToLower(strings.TrimSpace(parts[1]))
			}
			if len(parts) > 2 && strings.TrimSpace(parts[2]) != "" {
				rawState := strings.TrimSpace(parts[2])
				state = &rawState
			}
		}

		if portNumber == 0 {
			continue
		}
		if _, exists := seen[portNumber]; exists {
			continue
		}
		seen[portNumber] = struct{}{}
		ports = append(ports, TopologyPort{
			Number:   portNumber,
			Protocol: protocol,
			State:    state,
		})
	}

	return ports
}

func asStringSlice(value any) []string {
	items := asInterfaceSlice(value)
	if len(items) == 0 {
		if str, ok := value.(string); ok && str != "" {
			return []string{str}
		}
		return nil
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		if str := strings.TrimSpace(stringValue(item)); str != "" {
			result = append(result, str)
		}
	}
	return result
}

func asInterfaceSlice(value any) []any {
	switch raw := value.(type) {
	case []any:
		return raw
	default:
		return nil
	}
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstAny(values ...any) any {
	for _, value := range values {
		switch raw := value.(type) {
		case nil:
			continue
		case string:
			if strings.TrimSpace(raw) != "" {
				return raw
			}
		default:
			return value
		}
	}
	return nil
}

func stringValue(value any) string {
	switch raw := value.(type) {
	case nil:
		return ""
	case string:
		return raw
	case fmt.Stringer:
		return raw.String()
	case []byte:
		return string(raw)
	case json.Number:
		return raw.String()
	case float64:
		if raw == float64(int64(raw)) {
			return strconv.FormatInt(int64(raw), 10)
		}
		return strconv.FormatFloat(raw, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(raw), 'f', -1, 32)
	case int:
		return strconv.Itoa(raw)
	case int8:
		return strconv.FormatInt(int64(raw), 10)
	case int16:
		return strconv.FormatInt(int64(raw), 10)
	case int32:
		return strconv.FormatInt(int64(raw), 10)
	case int64:
		return strconv.FormatInt(raw, 10)
	case uint:
		return strconv.FormatUint(uint64(raw), 10)
	case uint8:
		return strconv.FormatUint(uint64(raw), 10)
	case uint16:
		return strconv.FormatUint(uint64(raw), 10)
	case uint32:
		return strconv.FormatUint(uint64(raw), 10)
	case uint64:
		return strconv.FormatUint(raw, 10)
	default:
		return fmt.Sprint(raw)
	}
}

func int32Value(value any) int32 {
	switch raw := value.(type) {
	case nil:
		return 0
	case int32:
		return raw
	case int:
		return int32(raw)
	case int64:
		return int32(raw)
	case float64:
		return int32(raw)
	case float32:
		return int32(raw)
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 32)
		return int32(n)
	default:
		n, _ := strconv.ParseInt(stringValue(raw), 10, 32)
		return int32(n)
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func (s *Neo4jTopologyService) executeQuery(ctx context.Context, cypher string, parameters map[string]any) ([][]any, error) {
	payload := neo4jTxRequest{
		Statements: []neo4jTxStatement{{
			Statement:  cypher,
			Parameters: parameters,
		}},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal topology query: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/db/%s/tx/commit", strings.TrimRight(s.url, "/"), s.database),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build topology query request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+basicAuth(s.user, s.password))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("neo4j topology query failed: %w", err)
	}
	defer resp.Body.Close()

	var txResp neo4jTxResponse
	if err := json.NewDecoder(resp.Body).Decode(&txResp); err != nil {
		return nil, fmt.Errorf("failed to decode neo4j topology response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("neo4j topology query failed with HTTP %d", resp.StatusCode)
	}
	if len(txResp.Errors) > 0 {
		messages := make([]string, 0, len(txResp.Errors))
		for _, item := range txResp.Errors {
			messages = append(messages, fmt.Sprintf("%s: %s", item.Code, item.Message))
		}
		return nil, fmt.Errorf("neo4j topology execution errors: %s", strings.Join(messages, "; "))
	}

	if len(txResp.Results) == 0 {
		return nil, nil
	}

	rows := make([][]any, 0, len(txResp.Results[0].Data))
	for _, item := range txResp.Results[0].Data {
		rows = append(rows, item.Row)
	}
	return rows, nil
}

func basicAuth(user, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, password)))
}
