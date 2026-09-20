import { Component } from "react";
import type { ErrorInfo, ReactNode } from "react";
import { Button } from "./ui";

interface Props {
  children: ReactNode;
}

interface State {
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("Unhandled render error", error, info.componentStack);
  }

  render() {
    if (this.state.error) {
      return (
        <div className="guard">
          <div className="error-splash">
            <h1>Something went wrong</h1>
            <p>{this.state.error.message || "Unexpected application error."}</p>
            <Button
              variant="secondary"
              onClick={() => {
                this.setState({ error: null });
                window.location.reload();
              }}
            >
              Reload
            </Button>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}