package app

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Derive a single cumulative patch from immutable, reviewed Git objects. Never
// package a caller-provided patch or the current (possibly dirty) checkout.
// A cumulative diff also includes accepted corrections from reopened tickets.
func approvedDeliveryPatch(ctx context.Context, plan RemediationPlan, state RemediationState) ([]byte, error) {
	if !commitPattern.MatchString(plan.BaselineCommit) || !commitPattern.MatchString(state.BestRef) || !plan.SecurityChangesOnly {
		return nil, fmt.Errorf("entrega exige commits imutáveis e alterações somente de segurança")
	}
	if _, err := canonicalRepository(plan.RepositoryPath); err != nil {
		return nil, fmt.Errorf("repositório aprovado indisponível")
	}
	if err := gitAncestor(plan.RepositoryPath, plan.BaselineCommit, state.BestRef); err != nil {
		return nil, fmt.Errorf("BEST não descende do baseline aprovado")
	}
	for _, ticket := range plan.Tickets {
		item := ticketState(&state, ticket.ID)
		if item == nil || (ticket.RequiresCodeChange && item.Status != "accepted") || (!ticket.RequiresCodeChange && item.Status != "no_code_final_review") {
			return nil, fmt.Errorf("ticket sem decisão final compatível")
		}
	}
	deadline, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	cmd := exec.CommandContext(deadline, "git", "-C", plan.RepositoryPath, "-c", "core.quotePath=true", "diff",
		"--no-ext-diff", "--no-textconv", "--binary", "--full-index", "--no-renames", "--no-color",
		"--src-prefix=a/", "--dst-prefix=b/", "--output-indicator-new=+", "--output-indicator-old=-", "--output-indicator-context= ",
		plan.BaselineCommit, state.BestRef, "--")
	output := &boundedBuffer{limit: 16 << 20}
	cmd.Stdout = output
	cmd.Env = remediationEnvironment(nil, nil)
	if err := cmd.Run(); err != nil || len(output.Bytes()) >= 16<<20 {
		return nil, fmt.Errorf("não foi possível derivar patch aprovado dentro dos limites")
	}
	if plan.BaselineCommit != state.BestRef && strings.TrimSpace(string(output.Bytes())) == "" {
		return nil, fmt.Errorf("commits diferentes sem correção exportável")
	}
	return output.Bytes(), nil
}
