import React from 'react';
import { Modal, Spin, Progress } from 'antd';
import { CloseOutlined } from '@ant-design/icons';
import useUiStore from '../store/uiStore';
import useProcessStore from '../store/processStore';

const LoadingModal = ({ onCancel }) => {
  const { showLoadingModal, loadingMessage } = useUiStore();
  const { currentTelemetry } = useProcessStore();

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 MB';
    const mb = bytes / (1024 * 1024);
    return mb.toFixed(1) + ' MB';
  };

  const calculateProgress = () => {
    if (!currentTelemetry || currentTelemetry.total_file_size === 0) {
      return 0;
    }
    return Math.round((currentTelemetry.bytes_read / currentTelemetry.total_file_size) * 100);
  };

  return (
    <Modal
      open={showLoadingModal}
      closable={true}
      closeIcon={<CloseOutlined />}
      onCancel={onCancel}
      footer={null}
      width={400}
      centered
      maskClosable={false}
    >
      <div id="loading-modal-content">
        {currentTelemetry ? (
          <>
            <Progress
              type="circle"
              percent={calculateProgress()}
              size={120}
            />
            <div id="loading-modal-progress-text">
              Processing: {formatBytes(currentTelemetry.bytes_read)} of {formatBytes(currentTelemetry.total_file_size)}
            </div>
            <div id="loading-modal-extracted-text">
              Extracted {currentTelemetry.persons_extracted.toLocaleString()} persons
            </div>
            {currentTelemetry.error_count > 0 && (
              <div id="loading-modal-error-text">
                {currentTelemetry.error_count.toLocaleString()} errors encountered
              </div>
            )}
          </>
        ) : (
          <>
            <Spin size="large" />
            <div id="loading-modal-message-text">
              {loadingMessage}
            </div>
          </>
        )}
      </div>
    </Modal>
  );
};

export default LoadingModal;