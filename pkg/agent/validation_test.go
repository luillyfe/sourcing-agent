package agent

import (
	"testing"
)

func TestRequirements_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     Requirements
		wantErr bool
	}{
		{
			name: "Valid Requirements",
			req: Requirements{
				RequiredSkills:  []string{"Go"},
				ExperienceLevel: "Senior",
			},
			wantErr: false,
		},
		{
			name: "Missing Skills",
			req: Requirements{
				RequiredSkills:  []string{},
				ExperienceLevel: "Senior",
			},
			wantErr: true,
		},
		{
			name: "Missing Experience",
			req: Requirements{
				RequiredSkills:  []string{"Go"},
				ExperienceLevel: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Requirements.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSearchStrategy_Validate(t *testing.T) {
	tests := []struct {
		name     string
		strategy SearchStrategy
		wantErr  bool
	}{
		{
			name: "Valid Strategy",
			strategy: SearchStrategy{
				PrimarySearch: SearchQuery{
					Language: "go",
				},
			},
			wantErr: false,
		},
		{
			name: "Missing Language",
			strategy: SearchStrategy{
				PrimarySearch: SearchQuery{
					Language: "",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.strategy.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("SearchStrategy.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnrichedCandidates_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cand    EnrichedCandidates
		wantErr bool
	}{
		{
			name: "Valid Candidates (Empty List)",
			cand: EnrichedCandidates{
				Candidates: []EnrichedCandidate{},
			},
			wantErr: false,
		},
		{
			name: "Valid Candidates (Populated)",
			cand: EnrichedCandidates{
				Candidates: []EnrichedCandidate{{Username: "test"}},
			},
			wantErr: false,
		},
		{
			name: "Invalid Candidates (Nil List)",
			cand: EnrichedCandidates{
				Candidates: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cand.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("EnrichedCandidates.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
