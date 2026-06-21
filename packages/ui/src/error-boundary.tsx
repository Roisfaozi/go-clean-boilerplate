import { AlertTriangle, RefreshCw } from "lucide-react";
import React from "react";
import { NexusButton } from "./nexus-button";

interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
  onReset?: () => void;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends React.Component<
  ErrorBoundaryProps,
  ErrorBoundaryState
> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error("[ErrorBoundary]", error, errorInfo);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null });
    this.props.onReset?.();
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) return this.props.fallback;

      return (
        <div className="flex min-h-[300px] flex-col items-center justify-center gap-4 p-8 text-center">
          <div className="bg-destructive/10 flex h-14 w-14 items-center justify-center rounded-full">
            <AlertTriangle className="text-destructive h-7 w-7" />
          </div>
          <div className="space-y-1">
            <h3 className="text-foreground text-lg font-semibold">
              Something went wrong
            </h3>
            <p className="text-muted-foreground max-w-md text-sm">
              {this.state.error?.message || "An unexpected error occurred."}
            </p>
          </div>
          <NexusButton variant="outline" onClick={this.handleReset}>
            <RefreshCw className="mr-2 h-4 w-4" />
            Try Again
          </NexusButton>
          {
            // @ts-ignore
            process.env.NODE_ENV !== "production" && this.state.error && (
              <pre className="bg-muted text-muted-foreground mt-4 max-w-2xl overflow-auto rounded-lg p-4 text-left text-xs">
                {this.state.error.stack}
              </pre>
            )
          }
        </div>
      );
    }

    return this.props.children;
  }
}
