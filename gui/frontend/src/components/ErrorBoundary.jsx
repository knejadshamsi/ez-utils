import React from 'react';
import { Result, Button } from 'antd';

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null });
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      return (
        <Result
          status="error"
          title="Something went wrong"
          subTitle={this.state.error?.message || 'An unexpected error occurred'}
          extra={[
            <Button type="primary" onClick={this.handleReset} key="reload">
              Reload Application
            </Button>,
            <pre key="stack" style={{ textAlign: 'left', whiteSpace: 'pre-wrap', background: '#fff0f0', border: '1px solid red', padding: '10px', marginTop: '10px' }}>
              {this.state.error?.stack}
            </pre>
          ]}
        />
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;