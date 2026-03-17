import { useEffect, useRef, useState } from 'react';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Checkbox from '@mui/material/Checkbox';
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import FormControlLabel from '@mui/material/FormControlLabel';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import StorageIcon from '@mui/icons-material/Storage';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';

import { importDatabaseMigration } from 'api/settings';
import { useTaskProgress } from 'contexts/TaskProgressContext';

export default function DatabaseMigrationSettings({ showMessage }) {
  const fileInputRef = useRef(null);
  const { registerOnComplete, unregisterOnComplete } = useTaskProgress();

  const [selectedFile, setSelectedFile] = useState(null);
  const [includeAccessKeys, setIncludeAccessKeys] = useState(true);
  const [includeSubLogs, setIncludeSubLogs] = useState(false);
  const [confirmOverwrite, setConfirmOverwrite] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [lastMigrationResult, setLastMigrationResult] = useState(null);

  useEffect(() => {
    const handleTaskComplete = ({ taskType, status, result }) => {
      if (taskType !== 'db_migration') {
        return;
      }

      if (status === 'completed') {
        const warningCount = result?.warnings?.length || 0;
        setLastMigrationResult(result || null);
        showMessage(
          warningCount > 0
            ? `迁移完成，但有 ${warningCount} 条警告。请重启项目实例后重新登录并检查数据`
            : '迁移完成，请重启项目实例后重新登录并检查数据',
          warningCount > 0 ? 'warning' : 'success'
        );
        return;
      }

      if (status === 'error') {
        showMessage('迁移任务失败，请到任务中心查看详情', 'error');
      }
    };

    registerOnComplete(handleTaskComplete);
    return () => {
      unregisterOnComplete(handleTaskComplete);
    };
  }, [registerOnComplete, showMessage, unregisterOnComplete]);

  const handleChooseFile = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (event) => {
    const file = event.target.files?.[0] || null;
    setSelectedFile(file);
  };

  const handleStartImport = async () => {
    if (!selectedFile) {
      showMessage('请先选择 backup.zip 或 SQLite 数据库文件', 'error');
      return;
    }

    if (!confirmOverwrite) {
      showMessage('请先确认本次导入会覆盖当前实例数据', 'error');
      return;
    }

    setSubmitting(true);
    try {
      const formData = new FormData();
      formData.append('file', selectedFile);
      formData.append('includeAccessKeys', String(includeAccessKeys));
      formData.append('includeSubLogs', String(includeSubLogs));

      setLastMigrationResult(null);
      await importDatabaseMigration(formData);
      showMessage('迁移任务已启动，可在右下角进度面板或任务中心查看进度');
    } catch (error) {
      showMessage(error.message || '启动迁移任务失败', 'error');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Card variant="outlined">
      <CardHeader
        avatar={<StorageIcon color="primary" />}
        title="SQLite 数据迁移"
        subheader="上传旧实例的 backup.zip 或 SQLite 数据库文件，导入到当前实例"
      />
      <CardContent>
        <Stack spacing={2.5}>
          <Alert
            severity="error"
            icon={<WarningAmberIcon fontSize="inherit" />}
            sx={{
              '& .MuiAlert-icon': {
                color: 'error.main'
              },
              '& .MuiAlert-message': {
                color: 'error.dark',
                fontWeight: 600
              }
            }}
          >
            本操作会覆盖当前实例的业务数据。建议只在新部署的 MySQL / PostgreSQL 实例首次迁移时使用，并先自行备份当前实例。
          </Alert>

          <Alert severity="info">
            推荐上传旧 SQLite 实例导出的 <strong>backup.zip</strong>。直接上传 <strong>.db</strong> 只会迁移数据库记录，不会恢复模板目录。
          </Alert>

          <Alert severity="info">
            旧实例导出备份方式：登录旧实例后台后，点击右上角头像菜单中的 <strong>系统备份</strong>，即可下载
            <strong> backup.zip </strong>
            文件。该压缩包会包含 <strong>db</strong> 和 <strong>template</strong> 目录，适合作为迁移源文件。
          </Alert>

          {lastMigrationResult?.warnings?.length > 0 && (
            <Alert severity="warning">
              <Stack spacing={1}>
                <Typography variant="body2" fontWeight={600}>
                  最近一次迁移包含 {lastMigrationResult.warnings.length} 条警告
                </Typography>
                <Box component="ul" sx={{ m: 0, pl: 2.5 }}>
                  {lastMigrationResult.warnings.map((warning, index) => (
                    <Box component="li" key={`${warning}-${index}`} sx={{ mb: 0.5 }}>
                      <Typography variant="body2">{warning}</Typography>
                    </Box>
                  ))}
                </Box>
                <Typography variant="caption" color="text.secondary">
                  这些警告也可在任务中心的“数据库迁移”任务详情中查看。
                </Typography>
              </Stack>
            </Alert>
          )}

          <Box>
            <input ref={fileInputRef} type="file" hidden accept=".zip,.db,.sqlite,.sqlite3" onChange={handleFileChange} />
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} alignItems={{ xs: 'stretch', sm: 'center' }}>
              <Button variant="outlined" startIcon={<CloudUploadIcon />} onClick={handleChooseFile} disabled={submitting}>
                选择迁移文件
              </Button>
              {selectedFile ? (
                <Chip
                  color="primary"
                  variant="outlined"
                  label={`${selectedFile.name} · ${(selectedFile.size / 1024 / 1024).toFixed(2)} MB`}
                  sx={{ maxWidth: '100%' }}
                />
              ) : (
                <Typography variant="body2" color="text.secondary">
                  支持 `backup.zip`、`.db`、`.sqlite`、`.sqlite3`
                </Typography>
              )}
            </Stack>
          </Box>

          <Divider />

          <Stack spacing={1}>
            <FormControlLabel
              control={<Checkbox checked={includeAccessKeys} onChange={(event) => setIncludeAccessKeys(event.target.checked)} />}
              label="迁移 AccessKey"
            />
            <Typography variant="caption" color="text.secondary" sx={{ ml: 4 }}>
              如果旧 API Key 迁移后无法使用，请确认当前实例的 API 加密密钥与旧实例一致，或在迁移后重新生成。
            </Typography>

            <FormControlLabel
              control={<Checkbox checked={includeSubLogs} onChange={(event) => setIncludeSubLogs(event.target.checked)} />}
              label="迁移订阅访问日志"
            />
            <Typography variant="caption" color="text.secondary" sx={{ ml: 4 }}>
              日志数据通常较大，默认不迁移。
            </Typography>

            <FormControlLabel
              control={
                <Checkbox checked={confirmOverwrite} onChange={(event) => setConfirmOverwrite(event.target.checked)} color="error" />
              }
              label={
                <Typography color="error.main" fontWeight={600}>
                  我已确认本次导入会覆盖当前实例的业务数据
                </Typography>
              }
            />
            <Typography variant="caption" color="error.main" sx={{ ml: 4, fontWeight: 600 }}>
              迁移任务完成后，请手动重启项目实例，再重新登录检查迁移结果。
            </Typography>
          </Stack>

          <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Button variant="contained" color="error" startIcon={<StorageIcon />} onClick={handleStartImport} disabled={submitting}>
              {submitting ? '提交中...' : '开始迁移'}
            </Button>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
}
