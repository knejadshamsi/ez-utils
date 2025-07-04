import React from 'react';
import { Modal, Spin, Progress } from 'antd';
import { CloseOutlined } from '@ant-design/icons';
import useAppStore from '../store/appStore';

const LoadingModal = ({ onCancel }) => {
  const { showLoadingModal, loadingMessage, currentTelemetry } = useAppStore();

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
      <div style={{ 
        display: 'flex', 
        flexDirection: 'column', 
        alignItems: 'center', 
        padding: '40px 20px' 
      }}>
        {currentTelemetry ? (
          <>
            <Progress 
              type="circle" 
              percent={calculateProgress()} 
              size={120}
            />
            <div style={{ marginTop: 24, fontSize: 16, textAlign: 'center' }}>
              Processing: {formatBytes(currentTelemetry.bytes_read)} of {formatBytes(currentTelemetry.total_file_size)}
            </div>
            <div style={{ marginTop: 8, fontSize: 14, color: '#666' }}>
              Extracted {currentTelemetry.persons_extracted.toLocaleString()} persons
            </div>
            {currentTelemetry.error_count > 0 && (
              <div style={{ marginTop: 8, fontSize: 14, color: '#ff4d4f' }}>
                {currentTelemetry.error_count.toLocaleString()} errors encountered
              </div>
            )}
          </>
        ) : (
          <>
            <Spin size="large" />
            <div style={{ marginTop: 24, fontSize: 16 }}>
              {loadingMessage}
            </div>
          </>
        )}
      </div>
    </Modal>
  );
};

export default LoadingModal;