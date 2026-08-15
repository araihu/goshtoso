package e2econstraints

import (
	"go/build/constraint"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateSuiteAcceptsSingleMultiAndSupportConstraints(t *testing.T) {
	suite := Suite{Files: []SourceFile{
		{Name: "support.go", Expr: mustConstraint(t, "//go:build e2e")},
		{Name: "button_test.go", Expr: mustConstraint(t, "//go:build e2e && (full || button)"), Tests: []string{"TestButton"}},
		{Name: "composition_test.go", Expr: mustConstraint(t, "//go:build e2e && (full || button || toolbar)"), Tests: []string{"TestComposition"}},
	}}
	manifest := Manifest{
		Identities: []Identity{
			{Name: "button", Files: []string{"button_test.go", "composition_test.go"}, Tests: []string{"TestButton", "TestComposition"}},
			{Name: "toolbar", Files: []string{"composition_test.go"}, Tests: []string{"TestComposition"}},
		},
		FullOnly: Identity{Name: "full_only"},
	}

	require.NoError(t, ValidateSuite(suite, manifest))
	require.Equal(t, []string{"TestButton", "TestComposition"}, suite.Inventory("button").Tests)
	require.Equal(t, []string{"TestComposition"}, suite.Inventory("toolbar").Tests)
}

func TestValidateSuiteAcceptsCurrentSourceOnlyRunnableConstraint(t *testing.T) {
	suite := Suite{Files: []SourceFile{
		{Name: "button_test.go", Expr: mustConstraint(t, "//go:build e2e && (full || button)"), Tests: []string{"TestButton"}},
		{Name: "theme_catalog_current_source_test.go", Expr: mustConstraint(t, "//go:build e2e && full && goshtoso_current_source"), Tests: []string{"TestThemeCatalog"}},
	}}
	manifest := Manifest{
		Identities: []Identity{
			{Name: "button", Files: []string{"button_test.go"}, Tests: []string{"TestButton"}},
		},
		FullOnly: Identity{Name: "full_only"},
	}

	require.NoError(t, ValidateSuite(suite, manifest))
	require.NotContains(t, suite.FullInventory().Tests, "TestThemeCatalog")
	require.Contains(t, suite.CurrentSourceInventory().Tests, "TestThemeCatalog")
}

func TestValidateSuiteAcceptsExplicitSpecializedSuite(t *testing.T) {
	suite := Suite{Files: []SourceFile{
		{Name: "support.go", Expr: mustConstraint(t, "//go:build e2e")},
		{Name: "button_test.go", Expr: mustConstraint(t, "//go:build e2e && (full || button)"), Tests: []string{"TestButton"}},
		{Name: "scrollregion_axe_test.go", Expr: mustConstraint(t, "//go:build e2e && scrollregion && axe"), Tests: []string{"TestScrollRegionAxe"}},
		{Name: "scrollregion_bfull_test.go", Expr: mustConstraint(t, "//go:build e2e && scrollregion && bfull && axe"), Tests: []string{"TestScrollRegionBFull"}},
	}}
	manifest := Manifest{
		Identities: []Identity{{Name: "button", Files: []string{"button_test.go"}, Tests: []string{"TestButton"}}},
		SpecializedSuites: []SpecializedSuite{
			{
				Name:          "scrollregion_axe",
				Tags:          []string{"scrollregion", "axe"},
				Files:         []string{"scrollregion_axe_test.go"},
				Tests:         []string{"TestScrollRegionAxe"},
				SelectedTests: []string{"TestScrollRegionAxe"},
			},
			{
				Name:          "scrollregion_bfull",
				Tags:          []string{"scrollregion", "bfull", "axe"},
				Files:         []string{"scrollregion_bfull_test.go"},
				Tests:         []string{"TestScrollRegionBFull"},
				SelectedTests: []string{"TestScrollRegionAxe", "TestScrollRegionBFull"},
			},
		},
		FullOnly: Identity{Name: "full_only"},
	}

	require.NoError(t, ValidateSuite(suite, manifest))
	require.Equal(t, []string{"TestScrollRegionBFull"}, suite.SpecializedInventory(manifest.SpecializedSuites[1]).Tests)
	var selections []string
	for _, selection := range suiteMatrixSelections(manifest) {
		selections = append(selections, strings.Join(selection.Tags, ","))
	}
	require.Equal(t, []string{"e2e,button", "e2e,scrollregion,axe", "e2e,scrollregion,bfull,axe", "e2e,full"}, selections)
}

func TestSpecializedSelectedTestsUsesManifestAsSingleRunnerAuthority(t *testing.T) {
	manifest := Manifest{SpecializedSuites: []SpecializedSuite{{
		Name:          "scrollregion_bfull",
		Tags:          []string{"scrollregion", "bfull", "axe"},
		Files:         []string{"scrollregion_bfull_test.go"},
		Tests:         []string{"TestScrollRegionBFull"},
		SelectedTests: []string{"TestScrollRegionBFull", "TestScrollRegionUA200HorizontalAccessContract"},
	}}}

	selected, err := SpecializedSelectedTests(manifest, "scrollregion_bfull")
	require.NoError(t, err)
	require.Equal(t, manifest.SpecializedSuites[0].SelectedTests, selected)
	selected[0] = "mutated"
	require.Equal(t, "TestScrollRegionBFull", manifest.SpecializedSuites[0].SelectedTests[0], "runner callers must not mutate manifest-owned selection")

	_, err = SpecializedSelectedTests(manifest, "missing")
	require.ErrorContains(t, err, `specialized suite "missing" is not declared`)
}

func TestValidateSuiteExcludesFullFallbackSpecializedSuiteFromFullOnly(t *testing.T) {
	suite := Suite{Files: []SourceFile{
		{Name: "backdrop_elevation_test.go", Expr: mustConstraint(t, "//go:build e2e && (full || backdropelevation)"), Tests: []string{"TestBackdropElevation"}},
	}}
	manifest := Manifest{
		SpecializedSuites: []SpecializedSuite{{
			Name:          "backdropelevation",
			Tags:          []string{"backdropelevation"},
			Files:         []string{"backdrop_elevation_test.go"},
			Tests:         []string{"TestBackdropElevation"},
			SelectedTests: []string{"TestBackdropElevation"},
		}},
		FullOnly: Identity{Name: "full_only"},
	}

	require.NoError(t, ValidateSuite(suite, manifest))
	require.Equal(t, []string{"backdrop_elevation_test.go"}, suite.SpecializedInventory(manifest.SpecializedSuites[0]).Files)
	require.Empty(t, suite.FullOnly(manifest).Tests)
}

func TestScrollRegionSpecializedSuitesOwnRealFiles(t *testing.T) {
	all, err := InspectSuite(filepath.Join(siteRoot(t), "tests", "e2e"))
	require.NoError(t, err)
	owned := map[string]bool{
		"scrollregion_axe_test.go":              true,
		"scrollregion_bfull_test.go":            true,
		"scrollregion_evidence_support_test.go": true,
		"scrollregion_evidence_test.go":         true,
		"scrollregion_test.go":                  true,
	}
	var suite Suite
	for _, file := range all.Files {
		if owned[file.Name] {
			suite.Files = append(suite.Files, file)
		}
	}
	manifest := Manifest{
		SpecializedSuites: []SpecializedSuite{
			{
				Name:  "scrollregion_axe",
				Tags:  []string{"scrollregion", "axe"},
				Files: []string{"scrollregion_axe_test.go"},
				Tests: []string{"TestScrollRegionAxeNamedPublicViewport"},
				SelectedTests: []string{
					"TestScrollRegionAxeNamedPublicViewport",
					"TestScrollRegionDirectRouteStatesAndInputModes",
				},
			},
			{
				Name:  "scrollregion_bfull",
				Tags:  []string{"scrollregion", "bfull", "axe"},
				Files: []string{"scrollregion_bfull_test.go", "scrollregion_evidence_support_test.go", "scrollregion_evidence_test.go"},
				Tests: []string{
					"TestScrollRegionATReceiptRejectsActionContractMutation",
					"TestScrollRegionATReceiptRejectsAttestationPayloadMutation",
					"TestScrollRegionATReceiptRejectsCandidateByteMutationBeforeReplayClaim",
					"TestScrollRegionATReceiptRejectsChallengeReplay",
					"TestScrollRegionATReceiptRejectsClaimantActionTimestampMismatch",
					"TestScrollRegionATReceiptRejectsGenericObservedSpeech",
					"TestScrollRegionATReceiptRejectsInferredFocusNavigation",
					"TestScrollRegionATReceiptRejectsNoOpHomeTransition",
					"TestScrollRegionATReceiptRejectsOverlappingActionWindow",
					"TestScrollRegionATReceiptRejectsPairKeyMismatch",
					"TestScrollRegionATReceiptRejectsPlaceholderOrZeroVersion",
					"TestScrollRegionATReceiptRejectsPlainTextScreenshot",
					"TestScrollRegionATReceiptRejectsReusedArtifacts",
					"TestScrollRegionATReceiptRejectsSecondReceiptForClaimedChallenge",
					"TestScrollRegionATReceiptRejectsSignedSyntheticCaptureWithoutRawProvenance",
					"TestScrollRegionATReceiptRejectsSignedSyntheticSolidPNGBeforeSemantics",
					"TestScrollRegionATReceiptRejectsStaleActionBoundVoiceOverLog",
					"TestScrollRegionATReceiptRejectsUnsignedCapture",
					"TestScrollRegionATReceiptRejectsWrongAttestationKey",
					"TestScrollRegionBFull",
					"TestScrollRegionBFullATReceiptHarness",
					"TestScrollRegionBFullFirstPaintObserverCannotMutateProductState",
					"TestScrollRegionBFullIdentityBindingRequiresSidecarForDirtyWorktree",
					"TestScrollRegionBFullIdentityBindingSealsVerifiedSidecar",
					"TestScrollRegionBFullPlanMarksExplicitDiagnosticNonClosure",
					"TestScrollRegionBFullPlanRejectsUnmarkedCap",
					"TestScrollRegionBFullReceiptRejectsLockedThemePersistenceClaim",
					"TestScrollRegionBFullReceiptRejectsRecomputedPersistenceForgeries",
					"TestScrollRegionBFullReceiptWrapperRejectsTampering",
					"TestScrollRegionBFullWrongChromeCropFailsVisualBinding",
					"TestScrollRegionCandidateIdentityRejectsDuplicatePath",
					"TestScrollRegionCandidateIdentityRejectsMissingDeclaredPath",
					"TestScrollRegionCandidateIdentityRejectsMutuallyConsistentFabrication",
					"TestScrollRegionCandidateIdentityRejectsTamperedDeclaredBytes",
					"TestScrollRegionCandidateIdentityRejectsUndeclaredExtraPath",
					"TestScrollRegionFooterUA200ResponsiveCompatibility",
					"TestScrollRegionReadAxeLockAllowsHumanReviewComments",
					"TestScrollRegionUA200HorizontalAccessContract",
				},
				SelectedTests: []string{
					"TestScrollRegionATReceiptRejectsActionContractMutation",
					"TestScrollRegionATReceiptRejectsAttestationPayloadMutation",
					"TestScrollRegionATReceiptRejectsCandidateByteMutationBeforeReplayClaim",
					"TestScrollRegionATReceiptRejectsChallengeReplay",
					"TestScrollRegionATReceiptRejectsClaimantActionTimestampMismatch",
					"TestScrollRegionATReceiptRejectsGenericObservedSpeech",
					"TestScrollRegionATReceiptRejectsInferredFocusNavigation",
					"TestScrollRegionATReceiptRejectsNoOpHomeTransition",
					"TestScrollRegionATReceiptRejectsOverlappingActionWindow",
					"TestScrollRegionATReceiptRejectsPairKeyMismatch",
					"TestScrollRegionATReceiptRejectsPlaceholderOrZeroVersion",
					"TestScrollRegionATReceiptRejectsPlainTextScreenshot",
					"TestScrollRegionATReceiptRejectsReusedArtifacts",
					"TestScrollRegionATReceiptRejectsSecondReceiptForClaimedChallenge",
					"TestScrollRegionATReceiptRejectsSignedSyntheticCaptureWithoutRawProvenance",
					"TestScrollRegionATReceiptRejectsSignedSyntheticSolidPNGBeforeSemantics",
					"TestScrollRegionATReceiptRejectsStaleActionBoundVoiceOverLog",
					"TestScrollRegionATReceiptRejectsUnsignedCapture",
					"TestScrollRegionATReceiptRejectsWrongAttestationKey",
					"TestScrollRegionAxeNamedPublicViewport",
					"TestScrollRegionBFull",
					"TestScrollRegionBFullATReceiptHarness",
					"TestScrollRegionBFullFirstPaintObserverCannotMutateProductState",
					"TestScrollRegionBFullIdentityBindingRequiresSidecarForDirtyWorktree",
					"TestScrollRegionBFullIdentityBindingSealsVerifiedSidecar",
					"TestScrollRegionBFullPlanMarksExplicitDiagnosticNonClosure",
					"TestScrollRegionBFullPlanRejectsUnmarkedCap",
					"TestScrollRegionBFullReceiptRejectsLockedThemePersistenceClaim",
					"TestScrollRegionBFullReceiptRejectsRecomputedPersistenceForgeries",
					"TestScrollRegionBFullReceiptWrapperRejectsTampering",
					"TestScrollRegionBFullWrongChromeCropFailsVisualBinding",
					"TestScrollRegionCandidateIdentityRejectsDuplicatePath",
					"TestScrollRegionCandidateIdentityRejectsMissingDeclaredPath",
					"TestScrollRegionCandidateIdentityRejectsMutuallyConsistentFabrication",
					"TestScrollRegionCandidateIdentityRejectsTamperedDeclaredBytes",
					"TestScrollRegionCandidateIdentityRejectsUndeclaredExtraPath",
					"TestScrollRegionDirectRouteStatesAndInputModes",
					"TestScrollRegionFooterUA200ResponsiveCompatibility",
					"TestScrollRegionReadAxeLockAllowsHumanReviewComments",
					"TestScrollRegionUA200HorizontalAccessContract",
				},
			},
		},
		FullOnly: Identity{Name: "full_only"},
	}
	require.NoError(t, ValidateSuite(suite, manifest))
}

func TestValidateSuiteRejectsUnsafeCurrentSourceConstraints(t *testing.T) {
	manifest := Manifest{Identities: []Identity{{Name: "button"}}}

	t.Run("leaks into standard full", func(t *testing.T) {
		suite := Suite{Files: []SourceFile{{
			Name:  "test.go",
			Expr:  mustConstraint(t, "//go:build e2e && (full || goshtoso_current_source)"),
			Tests: []string{"TestOne"},
		}}}
		err := ValidateSuite(suite, manifest)
		require.ErrorContains(t, err, "must be excluded from standard full")
	})

	t.Run("dedicated tags do not select test", func(t *testing.T) {
		suite := Suite{Files: []SourceFile{{
			Name:  "test.go",
			Expr:  mustConstraint(t, "//go:build e2e && !full && goshtoso_current_source"),
			Tests: []string{"TestOne"},
		}}}
		err := ValidateSuite(suite, manifest)
		require.ErrorContains(t, err, "missing dedicated tag inclusion")
	})

	t.Run("unknown tag remains rejected", func(t *testing.T) {
		suite := Suite{Files: []SourceFile{{
			Name:  "test.go",
			Expr:  mustConstraint(t, "//go:build e2e && full && goshtoso_current_source && typo"),
			Tests: []string{"TestOne"},
		}}}
		err := ValidateSuite(suite, manifest)
		require.ErrorContains(t, err, `unknown build tag "typo"`)
	})
}

func TestValidateSuiteRejectsUnsafeConstraints(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       string
	}{
		{name: "missing e2e", expression: "//go:build full || button", want: "does not require e2e"},
		{name: "missing full", expression: "//go:build e2e && button", want: "missing full fallback"},
		{name: "unknown", expression: "//go:build e2e && (full || typo)", want: "unknown build tag"},
		{name: "runnable support", expression: "//go:build e2e", want: "suite-only support"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			suite := Suite{Files: []SourceFile{{Name: "test.go", Expr: mustConstraint(t, test.expression), Tests: []string{"TestOne"}}}}
			err := ValidateSuite(suite, Manifest{Identities: []Identity{{Name: "button"}}})
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestLoadManifestRejectsDuplicateIdentities(t *testing.T) {
	path := t.TempDir() + "/identities.json"
	require.NoError(t, os.WriteFile(path, []byte(`{"identities":[{"name":"button"},{"name":"button"}]}`), 0o600))

	_, err := LoadManifest(path)
	require.ErrorContains(t, err, "duplicate identity")
}

func TestConstraintParserRejectsMalformedExpression(t *testing.T) {
	_, err := constraint.Parse("//go:build e2e && (")
	require.Error(t, err)
}

func mustConstraint(t *testing.T, expression string) constraint.Expr {
	t.Helper()
	parsed, err := constraint.Parse(expression)
	require.NoError(t, err)
	return parsed
}
