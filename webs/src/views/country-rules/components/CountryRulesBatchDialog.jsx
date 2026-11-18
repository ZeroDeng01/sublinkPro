import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import PropTypes from 'prop-types';

// material-ui
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import RadioGroup from '@mui/material/RadioGroup';
import FormControlLabel from '@mui/material/FormControlLabel';
import Radio from '@mui/material/Radio';
import Box from '@mui/material/Box';
import Alert from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import Collapse from '@mui/material/Collapse';
import IconButton from '@mui/material/IconButton';

// icons
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';

// ==============================|| BATCH EXAMPLE DATA ||============================== //

const BATCH_EXAMPLE = `# Country rules batch example
# Format: countryCode countryName pattern [priority] [status]

# Common regions
HK Hong Kong Hong Kong|HK|🇭🇰 100 enabled
TW Taiwan Taiwan|TW|Taipei 100 enabled
JP Japan Japan|JP|Tokyo|Osaka 90 enabled
SG Singapore Singapore|SG|Lion City 90 enabled
US United States "United States|US|USA|America|Los Angeles" 80 enabled

# Europe
GB United Kingdom "United Kingdom|GB|UK|Britain" 70 enabled
DE Germany Germany|DE|Frankfurt 70 enabled
FR France France|FR|Paris 70 disabled

# Priority and status are optional (default 0 and enabled)
CN China China|CN
KR South Korea South Korea|KR|Seoul`;

// ==============================|| PARSE UTILITIES ||============================== //

// 解析字段，支持引号
const parseFields = (line) => {
  const fields = [];
  let current = '';
  let inQuotes = false;

  for (let i = 0; i < line.length; i++) {
    const char = line[i];

    if (char === '"') {
      inQuotes = !inQuotes;
    } else if ((char === ' ' || char === '\t') && !inQuotes) {
      if (current) {
        fields.push(current);
        current = '';
      }
    } else {
      current += char;
    }
  }

  if (current) fields.push(current);
  return fields;
};

// 解析批量文本
const parseBatchText = (text, t) => {
  const lines = text.split('\n');
  const rules = [];
  const errors = [];
  let comments = 0;

  lines.forEach((line, index) => {
    const trimmed = line.trim();

    // 跳过空行和注释
    if (!trimmed || trimmed.startsWith('#')) {
      if (trimmed.startsWith('#')) comments++;
      return;
    }

    try {
      // 解析字段（支持引号）
      const fields = parseFields(trimmed);

      if (fields.length < 3) {
        errors.push({
          line: index + 1,
          message: t('settings.countryRulesPanel.batch.errors.insufficientFields')
        });
        return;
      }

      const rule = {
        countryCode: fields[0].toUpperCase(),
        countryName: fields[1],
        pattern: fields[2],
        priority: parseInt(fields[3]) || 0,
        enabled: fields[4] !== 'disabled'
      };

      // 验证国家代码
      if (!/^[A-Z]{2}$/.test(rule.countryCode)) {
        errors.push({
          line: index + 1,
          message: t('settings.countryRulesPanel.batch.errors.invalidCountryCode')
        });
        return;
      }

      // 验证正则表达式
      try {
        new RegExp(rule.pattern);
      } catch {
        errors.push({
          line: index + 1,
          message: t('settings.countryRulesPanel.batch.errors.invalidRegex')
        });
        return;
      }

      rules.push(rule);
    } catch (e) {
      errors.push({
        line: index + 1,
        message: e.message
      });
    }
  });

  return {
    totalLines: lines.length,
    validRules: rules.length,
    comments,
    errors,
    rules
  };
};

// ==============================|| COUNTRY RULES BATCH DIALOG ||============================== //

const CountryRulesBatchDialog = ({ open, onClose, onConfirm }) => {
  const { t } = useTranslation();
  const [batchText, setBatchText] = useState('');
  const [batchMode, setBatchMode] = useState('import');
  const [showFormat, setShowFormat] = useState(true);
  const [parseResult, setParseResult] = useState({
    totalLines: 0,
    validRules: 0,
    comments: 0,
    errors: []
  });

  // 实时解析文本
  useEffect(() => {
    if (batchText) {
      const result = parseBatchText(batchText, t);
      setParseResult(result);
    } else {
      setParseResult({
        totalLines: 0,
        validRules: 0,
        comments: 0,
        errors: []
      });
    }
  }, [batchText, t]);

  // 处理确认
  const handleConfirm = () => {
    if (parseResult.errors.length > 0) {
      return;
    }

    if (parseResult.validRules === 0) {
      return;
    }

    // 覆盖模式需要二次确认
    if (batchMode === 'replace') {
      if (!window.confirm(t('settings.countryRulesPanel.batch.messages.confirmReplace'))) {
        return;
      }
    }

    onConfirm({
      mode: batchMode,
      rules: parseResult.rules
    });
  };

  // 加载示例
  const handleLoadExample = () => {
    setBatchText(BATCH_EXAMPLE);
  };

  // 清空
  const handleClear = () => {
    setBatchText('');
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>{t('settings.countryRulesPanel.batch.title')}</DialogTitle>
      <DialogContent>
        {/* 操作模式选择 */}
        <Box sx={{ mb: 2 }}>
          <RadioGroup row value={batchMode} onChange={(e) => setBatchMode(e.target.value)}>
            <FormControlLabel
              value="import"
              control={<Radio />}
              label={
                <Box>
                  <Typography variant="body2" fontWeight="bold">
                    {t('settings.countryRulesPanel.batch.mode.import')}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    {t('settings.countryRulesPanel.batch.mode.importDesc')}
                  </Typography>
                </Box>
              }
            />
            <FormControlLabel
              value="replace"
              control={<Radio />}
              label={
                <Box>
                  <Typography variant="body2" fontWeight="bold">
                    {t('settings.countryRulesPanel.batch.mode.replace')}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    {t('settings.countryRulesPanel.batch.mode.replaceDesc')}
                  </Typography>
                </Box>
              }
            />
          </RadioGroup>
        </Box>

        {/* 格式说明（可折叠） */}
        <Box sx={{ mb: 2 }}>
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              cursor: 'pointer',
              mb: 1
            }}
            onClick={() => setShowFormat(!showFormat)}
          >
            <Typography variant="subtitle2" sx={{ flexGrow: 1 }}>
              {t('settings.countryRulesPanel.batch.format.title')}
            </Typography>
            <IconButton size="small">{showFormat ? <ExpandLessIcon /> : <ExpandMoreIcon />}</IconButton>
          </Box>
          <Collapse in={showFormat}>
            <Alert severity="info" sx={{ '& .MuiAlert-message': { width: '100%' } }}>
              <Typography variant="body2" sx={{ mb: 1 }}>
                {t('settings.countryRulesPanel.batch.format.description')}
              </Typography>
              <Typography variant="caption" component="pre" sx={{ fontFamily: 'monospace', whiteSpace: 'pre-wrap' }}>
                {t('settings.countryRulesPanel.batch.format.example')}
              </Typography>
            </Alert>
          </Collapse>
        </Box>

        {/* 文本编辑器 */}
        <TextField
          fullWidth
          multiline
          rows={12}
          value={batchText}
          onChange={(e) => setBatchText(e.target.value)}
          placeholder={t('settings.countryRulesPanel.batch.textField.placeholder')}
          sx={{
            mb: 2,
            '& .MuiInputBase-root': {
              fontFamily: 'monospace',
              fontSize: '0.875rem'
            }
          }}
        />

        {/* 解析统计 */}
        {batchText && (
          <Box sx={{ mb: 2 }}>
            <Typography variant="subtitle2" gutterBottom>
              📊 {t('settings.countryRulesPanel.batch.stats.title')}
            </Typography>
            <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap' }}>
              <Typography variant="body2">
                • {t('settings.countryRulesPanel.batch.stats.totalLines')}: {parseResult.totalLines}
              </Typography>
              <Typography variant="body2" color="success.main">
                • {t('settings.countryRulesPanel.batch.stats.validRules')}: {parseResult.validRules}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                • {t('settings.countryRulesPanel.batch.stats.comments')}: {parseResult.comments}
              </Typography>
              <Typography variant="body2" color={parseResult.errors.length > 0 ? 'error.main' : 'text.secondary'}>
                • {t('settings.countryRulesPanel.batch.stats.errors')}: {parseResult.errors.length}
              </Typography>
            </Box>
          </Box>
        )}

        {/* 错误列表 */}
        {parseResult.errors.length > 0 && (
          <Alert severity="error" sx={{ mb: 2 }}>
            <Typography variant="body2" fontWeight="bold" gutterBottom>
              {t('settings.countryRulesPanel.batch.errors.title', { count: parseResult.errors.length })}
            </Typography>
            <Box component="ul" sx={{ mt: 1, mb: 0, pl: 2 }}>
              {parseResult.errors.slice(0, 5).map((err, idx) => (
                <li key={idx}>{t('settings.countryRulesPanel.batch.errors.lineError', { line: err.line, message: err.message })}</li>
              ))}
              {parseResult.errors.length > 5 && (
                <li>{t('settings.countryRulesPanel.batch.errors.moreErrors', { count: parseResult.errors.length - 5 })}</li>
              )}
            </Box>
          </Alert>
        )}
      </DialogContent>

      <DialogActions>
        <Button onClick={onClose}>{t('settings.countryRulesPanel.batch.actions.cancel')}</Button>
        <Button onClick={handleClear} color="secondary">
          {t('settings.countryRulesPanel.batch.actions.clear')}
        </Button>
        <Button onClick={handleLoadExample} color="info">
          {t('settings.countryRulesPanel.batch.actions.example')}
        </Button>
        <Button onClick={handleConfirm} variant="contained" disabled={parseResult.errors.length > 0 || parseResult.validRules === 0}>
          {t('settings.countryRulesPanel.batch.actions.confirm')}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

CountryRulesBatchDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onConfirm: PropTypes.func.isRequired
};

export default CountryRulesBatchDialog;
