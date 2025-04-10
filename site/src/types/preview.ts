// Code generated by 'guts'. DO NOT EDIT.

// From types/diagnostics.go
export type DiagnosticSeverityString = "error" | "warning";

export const DiagnosticSeverityStrings: DiagnosticSeverityString[] = ["error", "warning"];

// From types/diagnostics.go
export type Diagnostics = readonly (FriendlyDiagnostic)[];

// From types/diagnostics.go
export interface FriendlyDiagnostic {
    readonly severity: DiagnosticSeverityString;
    readonly summary: string;
    readonly detail: string;
}

// From types/value.go
export interface NullHCLString {
    readonly value: string;
    readonly valid: boolean;
}

// From types/parameter.go
export interface Parameter extends ParameterData {
    readonly value: NullHCLString;
    readonly diagnostics: Diagnostics;
}

// From types/parameter.go
export interface ParameterData {
    readonly name: string;
    readonly display_name: string;
    readonly description: string;
    readonly type: ParameterType;
    // this is likely an enum in an external package "github.com/coder/terraform-provider-coder/v2/provider.ParameterFormType"
    readonly form_type: string;
    // empty interface{} type, falling back to unknown
    readonly styling: unknown;
    readonly mutable: boolean;
    readonly default_value: NullHCLString;
    readonly icon: string;
    readonly options: readonly (ParameterOption)[];
    readonly validations: readonly (ParameterValidation)[];
    readonly required: boolean;
    readonly order: number;
    readonly ephemeral: boolean;
}

// From types/parameter.go
export interface ParameterOption {
    readonly name: string;
    readonly description: string;
    readonly value: NullHCLString;
    readonly icon: string;
}

// From types/enum.go
export type ParameterType = "bool" | "list(string)" | "number" | "string";

export const ParameterTypes: ParameterType[] = ["bool", "list(string)", "number", "string"];

// From types/parameter.go
export interface ParameterValidation {
    readonly validation_error: string;
    readonly validation_regex: string | null;
    readonly validation_min: number | null;
    readonly validation_max: number | null;
    readonly validation_monotonic: string | null;
}

// From web/session.go
export interface Request {
    readonly id: number;
    readonly inputs: Record<string, string>;
}

// From web/session.go
export interface Response {
    readonly id: number;
    readonly diagnostics: Diagnostics;
    readonly parameters: readonly Parameter[];
}

// From web/session.go
export interface SessionInputs {
    readonly PlanPath: string;
    readonly User: WorkspaceOwner;
}

// From types/parameter.go
export const ValidationMonotonicDecreasing = "decreasing";

// From types/parameter.go
export const ValidationMonotonicIncreasing = "increasing";

// From types/owner.go
export interface WorkspaceOwner {
    readonly groups: readonly string[];
}

