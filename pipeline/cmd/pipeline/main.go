package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"empresa-security/pipeline/internal/app"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var err error
	switch os.Args[1] {
	case "pipeline1":
		err = pipeline1(ctx, os.Args[2:])
	case "pipeline2":
		err = pipeline2(ctx, os.Args[2:])
	case "verify":
		err = verify(os.Args[2:])
	case "tools":
		err = tools(os.Args[2:])
	case "normalize":
		err = normalize(os.Args[2:])
	case "consolidate":
		err = consolidate(os.Args[2:])
	case "compare":
		err = compare(os.Args[2:])
	case "supply-chain":
		err = supplyChain(os.Args[2:])
	case "runtime-check":
		err = runtimeCheck(ctx, os.Args[2:])
	case "audit-proxy":
		err = auditProxy(ctx, os.Args[2:])
	case "isolation-smoke":
		err = isolationSmoke(ctx, os.Args[2:])
	case "remediation-plan":
		err = remediationPlan(os.Args[2:])
	case "remediation-preflight":
		err = remediationPreflight(os.Args[2:])
	case "remediation-worktree":
		err = remediationWorktree(os.Args[2:])
	case "remediation-agent":
		err = remediationAgent(ctx, os.Args[2:])
	case "remediation-gates":
		err = remediationGates(ctx, os.Args[2:])
	case "remediation-evaluate":
		err = remediationEvaluate(os.Args[2:])
	case "remediation-finalize":
		err = remediationFinalize(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

func remediationPreflight(args []string) error {
	fs := flag.NewFlagSet("remediation-preflight", flag.ContinueOnError)
	configPath := fs.String("config", "", "política JSON de correção")
	outputPath := fs.String("output", "remediation-readiness.json", "relatório JSON de prontidão")
	strict := fs.Bool("strict", false, "falhar se qualquer adaptador estiver bloqueado")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" || *outputPath == "" {
		return fmt.Errorf("--config e --output são obrigatórios")
	}
	cfg, err := app.ReadRemediationConfig(*configPath)
	if err != nil {
		return err
	}
	report, err := app.CheckRemediationReadiness(*configPath, cfg)
	if err != nil {
		return err
	}
	if err := app.WriteJSON(*outputPath, report); err != nil {
		return err
	}
	for _, check := range report.Checks {
		fmt.Printf("%-7s %-24s %-8s %s\n", check.Kind, check.ID, check.Status, check.Reason)
	}
	if *strict && !report.Approved {
		return fmt.Errorf("adaptadores de correção não estão prontos")
	}
	return nil
}

func remediationPlan(args []string) error {
	fs := flag.NewFlagSet("remediation-plan", flag.ContinueOnError)
	configPath := fs.String("config", "", "política JSON de correção")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" {
		return fmt.Errorf("--config é obrigatório")
	}
	cfg, err := app.ReadRemediationConfig(*configPath)
	if err != nil {
		return err
	}
	planPath, err := app.CreateRemediationPlan(*configPath, cfg)
	if err == nil {
		fmt.Println(planPath)
	}
	return err
}

func remediationWorktree(args []string) error {
	fs := flag.NewFlagSet("remediation-worktree", flag.ContinueOnError)
	configPath := fs.String("config", "", "política JSON de correção")
	planPath := fs.String("plan", "", "plan.json imutável")
	ticketID := fs.String("ticket", "", "ticket REM-*")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" || *planPath == "" || *ticketID == "" {
		return fmt.Errorf("--config, --plan e --ticket são obrigatórios")
	}
	cfg, err := app.ReadRemediationConfig(*configPath)
	if err != nil {
		return err
	}
	metadataPath, err := app.PrepareRemediationWorktree(cfg, *planPath, *ticketID)
	if err == nil {
		fmt.Println(metadataPath)
	}
	return err
}

func remediationAgent(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("remediation-agent", flag.ContinueOnError)
	configPath := fs.String("config", "", "política JSON de correção")
	planPath := fs.String("plan", "", "plan.json imutável")
	ticketID := fs.String("ticket", "", "ticket REM-*")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" || *planPath == "" || *ticketID == "" {
		return fmt.Errorf("--config, --plan e --ticket são obrigatórios")
	}
	cfg, err := app.ReadRemediationConfig(*configPath)
	if err != nil {
		return err
	}
	runPath, err := app.RunRemediationAgent(ctx, cfg, *planPath, *ticketID)
	if runPath != "" {
		fmt.Println(runPath)
	}
	return err
}

func remediationGates(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("remediation-gates", flag.ContinueOnError)
	configPath := fs.String("config", "", "política JSON de correção")
	planPath := fs.String("plan", "", "plan.json imutável")
	ticketID := fs.String("ticket", "", "ticket REM-*; omita somente no gate global")
	phase := fs.String("phase", "candidate", "candidate ou global")
	candidateRef := fs.String("candidate-ref", "", "ref/commit Git em checkout")
	worktreePath := fs.String("worktree", "", "raiz do worktree preparado; omita para usar repository_path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" || *planPath == "" || *candidateRef == "" {
		return fmt.Errorf("--config, --plan e --candidate-ref são obrigatórios")
	}
	if *phase == "candidate" && *ticketID == "" {
		return fmt.Errorf("--ticket é obrigatório na fase candidate")
	}
	cfg, err := app.ReadRemediationConfig(*configPath)
	if err != nil {
		return err
	}
	runPath, err := app.RunRemediationGates(ctx, cfg, *planPath, *ticketID, *phase, *candidateRef, *worktreePath)
	if err == nil {
		fmt.Println(runPath)
	}
	return err
}

func remediationEvaluate(args []string) error {
	fs := flag.NewFlagSet("remediation-evaluate", flag.ContinueOnError)
	configPath := fs.String("config", "", "política JSON de correção")
	planPath := fs.String("plan", "", "plan.json imutável")
	ticketID := fs.String("ticket", "", "ticket REM-*; omita somente com --global")
	candidateRef := fs.String("candidate-ref", "", "ref/commit Git avaliado")
	candidateBundle := fs.String("candidate-bundle", "", "bundle canônico produzido pelo gate security")
	gateRun := fs.String("gate-run", "", "gate-run.json correspondente")
	global := fs.Bool("global", false, "avaliar gate global da melhor versão")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" || *planPath == "" || *candidateRef == "" || *candidateBundle == "" || *gateRun == "" {
		return fmt.Errorf("--config, --plan, --candidate-ref, --candidate-bundle e --gate-run são obrigatórios")
	}
	if !*global && *ticketID == "" {
		return fmt.Errorf("--ticket é obrigatório sem --global")
	}
	cfg, err := app.ReadRemediationConfig(*configPath)
	if err != nil {
		return err
	}
	evaluationPath, err := app.EvaluateRemediationCandidate(cfg, *planPath, *ticketID, *candidateRef, *candidateBundle, *gateRun, *global)
	if err == nil {
		fmt.Println(evaluationPath)
	}
	return err
}

func remediationFinalize(args []string) error {
	fs := flag.NewFlagSet("remediation-finalize", flag.ContinueOnError)
	configPath := fs.String("config", "", "política JSON de correção")
	planPath := fs.String("plan", "", "plan.json imutável")
	approvalPath := fs.String("approval", "", "decisão humana final JSON")
	outputPath := fs.String("output", "", "autorização de entrega JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" || *planPath == "" || *approvalPath == "" || *outputPath == "" {
		return fmt.Errorf("--config, --plan, --approval e --output são obrigatórios")
	}
	cfg, err := app.ReadRemediationConfig(*configPath)
	if err != nil {
		return err
	}
	return app.FinalizeRemediation(*configPath, cfg, *planPath, *approvalPath, *outputPath)
}

func isolationSmoke(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("isolation-smoke", flag.ContinueOnError)
	policyPath := fs.String("engagement-policy", "", "engagement-policy.json para infraestrutura controlada pelo operador")
	caseDir := fs.String("case-dir", "", "diretório exclusivo para os artefatos do teste")
	maxRequests := fs.Int("max-requests", 1, "budget do único request pelo proxy")
	rps := fs.Float64("requests-per-second", 1, "taxa máxima do request pelo proxy")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *policyPath == "" || *caseDir == "" {
		return fmt.Errorf("--engagement-policy e --case-dir são obrigatórios")
	}
	policy, err := app.ReadEngagementPolicy(*policyPath)
	if err != nil {
		return err
	}
	result, err := app.RunIsolationSmoke(ctx, policy, *caseDir, *maxRequests, *rps)
	if err != nil {
		return err
	}
	return app.WriteJSON(*caseDir+string(os.PathSeparator)+"isolation-smoke-result.json", result)
}

func auditProxy(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("audit-proxy", flag.ContinueOnError)
	policyPath := fs.String("engagement-policy", "", "engagement-policy.json validado")
	httpAddress := fs.String("http", "127.0.0.1:0", "endereço do proxy HTTP")
	socksAddress := fs.String("socks5", "127.0.0.1:0", "endereço do proxy SOCKS5")
	logPath := fs.String("log", "requestlog.ndjson", "RequestLog NDJSON")
	maxRequests := fs.Int("max-requests", 10000, "budget máximo do proxy")
	rps := fs.Float64("requests-per-second", 50, "taxa máxima agregada")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *policyPath == "" {
		return fmt.Errorf("--engagement-policy é obrigatório")
	}
	policy, err := app.ReadEngagementPolicy(*policyPath)
	if err != nil {
		return err
	}
	transport := app.NewPolicyScopeTransport(policy, *maxRequests, *rps)
	proxy, err := app.StartAuditProxy(ctx, *httpAddress, *socksAddress, *logPath, transport)
	if err != nil {
		return err
	}
	fmt.Printf("HTTP_PROXY=http://%s\nALL_PROXY=socks5://%s\n", proxy.HTTPAddress, proxy.SOCKSAddress)
	<-ctx.Done()
	return proxy.Close(context.Background())
}

func pipeline1(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("pipeline1", flag.ContinueOnError)
	config := fs.String("config", "", "arquivo JSON de entrada do lead")
	execute := fs.Bool("execute", false, "executar coleta; sem esta opção apenas planeja")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *config == "" {
		return fmt.Errorf("--config é obrigatório")
	}
	var cfg app.LeadConfig
	if err := app.ReadJSONWithSchema(*config, "lead.schema.json", &cfg); err != nil {
		return err
	}
	return app.RunPipeline1(ctx, cfg, *execute)
}

func pipeline2(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("pipeline2", flag.ContinueOnError)
	config := fs.String("config", "", "arquivo JSON de escopo autorizado")
	execute := fs.Bool("execute", false, "executar módulos autorizados; sem esta opção apenas planeja")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *config == "" {
		return fmt.Errorf("--config é obrigatório")
	}
	var cfg app.ScopeConfig
	if err := app.ReadJSONWithSchema(*config, "scope.schema.json", &cfg); err != nil {
		return err
	}
	return app.RunPipeline2(ctx, cfg, *execute)
}

func verify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	dir := fs.String("case-dir", "", "diretório do caso")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dir == "" {
		return fmt.Errorf("--case-dir é obrigatório")
	}
	return app.VerifyCase(*dir)
}

func tools(args []string) error {
	fs := flag.NewFlagSet("tools", flag.ContinueOnError)
	timeout := fs.Duration("timeout", 4*time.Second, "timeout por consulta de versão")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for _, status := range app.ToolStatuses(*timeout) {
		fmt.Printf("%-14s %-10s %s\n", status.Name, status.State, status.Version)
	}
	return nil
}

type repeatedFlag []string

func (value *repeatedFlag) String() string        { return fmt.Sprint([]string(*value)) }
func (value *repeatedFlag) Set(item string) error { *value = append(*value, item); return nil }

func normalize(args []string) error {
	fs := flag.NewFlagSet("normalize", flag.ContinueOnError)
	tool := fs.String("tool", "", "ferramenta que gerou a saída")
	input := fs.String("input", "", "arquivo bruto JSON, JSONL, SARIF, XML ou CSV")
	output := fs.String("output", "", "bundle canônico JSON")
	engagement := fs.String("engagement", "", "identificador do engagement")
	scope := fs.String("scope", "", "identificador estável do escopo")
	version := fs.String("tool-version", "", "versão observada da ferramenta")
	supply := fs.String("supply-chain-status", "unknown", "estado da cadeia da execução importada")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *tool == "" || *input == "" || *output == "" || *engagement == "" || *scope == "" {
		return fmt.Errorf("--tool, --input, --output, --engagement e --scope são obrigatórios")
	}
	bundle, err := app.NormalizeFile(app.NormalizeConfig{Tool: *tool, InputPath: *input, EngagementID: *engagement, ScopeID: *scope, ToolVersion: *version, SupplyChainStatus: *supply})
	if err != nil {
		return err
	}
	return app.WriteBundle(*output, bundle)
}

func consolidate(args []string) error {
	fs := flag.NewFlagSet("consolidate", flag.ContinueOnError)
	var inputs repeatedFlag
	fs.Var(&inputs, "input", "bundle canônico; pode ser repetido")
	output := fs.String("output", "", "bundle consolidado JSON")
	report := fs.String("report", "", "relatório Markdown opcional")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(inputs) == 0 || *output == "" {
		return fmt.Errorf("ao menos um --input e --output são obrigatórios")
	}
	bundles := make([]app.NormalizedBundle, 0, len(inputs))
	for _, input := range inputs {
		bundle, err := app.ReadBundle(input)
		if err != nil {
			return err
		}
		bundles = append(bundles, bundle)
	}
	merged, err := app.MergeBundles(bundles...)
	if err != nil {
		return err
	}
	if err := app.WriteBundle(*output, merged); err != nil {
		return err
	}
	if *report != "" {
		return app.WriteNormalizedReport(*report, merged)
	}
	return nil
}

func compare(args []string) error {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	previous := fs.String("previous", "", "bundle anterior")
	current := fs.String("current", "", "bundle atual")
	output := fs.String("output", "", "comparação JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *previous == "" || *current == "" || *output == "" {
		return fmt.Errorf("--previous, --current e --output são obrigatórios")
	}
	oldBundle, err := app.ReadBundle(*previous)
	if err != nil {
		return err
	}
	newBundle, err := app.ReadBundle(*current)
	if err != nil {
		return err
	}
	return app.WriteComparison(*output, app.CompareBundles(oldBundle, newBundle))
}

func supplyChain(args []string) error {
	fs := flag.NewFlagSet("supply-chain", flag.ContinueOnError)
	lockPath := fs.String("lock", "tools.lock.json", "manifesto de ferramentas")
	sbomPath := fs.String("sbom", "sbom.cdx.json", "SBOM CycloneDX")
	output := fs.String("output", "", "relatório JSON opcional")
	strict := fs.Bool("strict", false, "falhar se houver qualquer bloqueio")
	if err := fs.Parse(args); err != nil {
		return err
	}
	lock, err := app.ReadToolLock(*lockPath)
	if err != nil {
		return err
	}
	report := app.ValidateToolLock(lock, *sbomPath)
	report.LockPath = *lockPath
	if *output != "" {
		if err := app.WriteJSON(*output, report); err != nil {
			return err
		}
	}
	for _, check := range report.Checks {
		state := "approved"
		if !check.Approved {
			state = "blocked"
		} else if check.ReviewStatus == "restricted" {
			state = "restricted"
		}
		fmt.Printf("%-16s %-9s %s\n", check.Tool, state, strings.Join(check.Blockers, "; "))
	}
	if *strict && !report.Approved {
		return fmt.Errorf("cadeia de ferramentas possui bloqueios")
	}
	return nil
}

func runtimeCheck(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("runtime-check", flag.ContinueOnError)
	lockPath := fs.String("lock", "tools.lock.json", "manifesto de ferramentas")
	sbomPath := fs.String("sbom", "sbom.cdx.json", "SBOM CycloneDX")
	output := fs.String("output", "runtime-readiness.json", "relatório JSON")
	pull := fs.Bool("pull", false, "baixar previamente cada imagem pelo digest exato")
	strict := fs.Bool("strict", false, "falhar se qualquer runtime automático não passar")
	if err := fs.Parse(args); err != nil {
		return err
	}
	lock, err := app.ReadToolLock(*lockPath)
	if err != nil {
		return err
	}
	report, err := app.CheckRuntimeReadiness(ctx, lock, *sbomPath, *pull)
	if err != nil {
		return err
	}
	if err := app.WriteJSON(*output, report); err != nil {
		return err
	}
	for _, check := range report.Checks {
		fmt.Printf("%-20s %-10s %s\n", check.Tool, check.Status, check.Reason)
	}
	if *strict && !report.Approved {
		return fmt.Errorf("prontidão de runtime possui falhas")
	}
	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "uso: pipeline <pipeline1|pipeline2|normalize|consolidate|compare|supply-chain|runtime-check|audit-proxy|isolation-smoke|remediation-preflight|remediation-plan|remediation-worktree|remediation-agent|remediation-gates|remediation-evaluate|remediation-finalize|verify|tools> [opções]")
}
