import { useMemo } from "react";
import useMediaQuery from "@mui/material/useMediaQuery";
import { useTheme } from "@mui/material/styles";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import TextField from "@mui/material/TextField";
import Stack from "@mui/material/Stack";
import Alert from "@mui/material/Alert";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import Checkbox from "@mui/material/Checkbox";
import FormControlLabel from "@mui/material/FormControlLabel";
import Radio from "@mui/material/Radio";
import RadioGroup from "@mui/material/RadioGroup";
import Typography from "@mui/material/Typography";
import Autocomplete from "@mui/material/Autocomplete";
import Tooltip from "@mui/material/Tooltip";
import InputAdornment from "@mui/material/InputAdornment";
import Divider from "@mui/material/Divider";
import Grid from "@mui/material/Grid";
import ButtonGroup from "@mui/material/ButtonGroup";
import BuildIcon from "@mui/icons-material/Build";
import EditNoteIcon from "@mui/icons-material/EditNote";

import NodeRenameBuilder from "./NodeRenameBuilder";
import NodeNamePreprocessor from "./NodeNamePreprocessor";
import NodeNameFilter from "./NodeNameFilter";
import NodeTransferBox from "./NodeTransferBox";

// ISO国家代码转换为国旗emoji
const isoToFlag = (isoCode) => {
  if (!isoCode || isoCode.length !== 2) return "";
  const code = isoCode.toUpperCase() === "TW" ? "CN" : isoCode.toUpperCase();
  const codePoints = code.split("").map((char) => 0x1f1e6 + char.charCodeAt(0) - 65);
  return String.fromCodePoint(...codePoints);
};

// 格式化国家显示
const formatCountry = (linkCountry) => {
  if (!linkCountry) return "";
  const flag = isoToFlag(linkCountry);
  return flag ? `${flag}${linkCountry}` : linkCountry;
};

// 预览节点名称
const previewNodeName = (rule) => {
  if (!rule) return "";
  return rule
    .replace(/\$Name/g, "香港节点-备注")
    .replace(/\$Flag/g, "🇭🇰")
    .replace(/\$LinkName/g, "香港01")
    .replace(/\$LinkCountry/g, "HK")
    .replace(/\$Speed/g, "1.50MB/s")
    .replace(/\$Delay/g, "125ms")
    .replace(/\$Group/g, "Premium")
    .replace(/\$Source/g, "机场A")
    .replace(/\$Index/g, "1")
    .replace(/\$Protocol/g, "VMess");
};

/**
 * 订阅表单对话框
 */
export default function SubscriptionFormDialog({
                                                 open,
                                                 isEdit,
                                                 formData,
                                                 setFormData,
                                                 templates,
                                                 scripts,
                                                 allNodes,
                                                 groupOptions,
                                                 sourceOptions,
                                                 countryOptions,
                                                 // 节点过滤
                                                 nodeGroupFilter,
                                                 setNodeGroupFilter,
                                                 nodeSourceFilter,
                                                 setNodeSourceFilter,
                                                 nodeSearchQuery,
                                                 setNodeSearchQuery,
                                                 nodeCountryFilter,
                                                 setNodeCountryFilter,
                                                 // 穿梭框状态
                                                 checkedAvailable,
                                                 checkedSelected,
                                                 mobileTab,
                                                 setMobileTab,
                                                 selectedNodeSearch,
                                                 setSelectedNodeSearch,
                                                 namingMode,
                                                 setNamingMode,
                                                 // 操作回调
                                                 onClose,
                                                 onSubmit,
                                                 onAddNode,
                                                 onRemoveNode,
                                                 onAddAllVisible,
                                                 onRemoveAll,
                                                 onToggleAvailable,
                                                 onToggleSelected,
                                                 onAddChecked,
                                                 onRemoveChecked,
                                                 onToggleAllAvailable,
                                                 onToggleAllSelected
                                               }) {
  const theme = useTheme();
  const matchDownMd = useMediaQuery(theme.breakpoints.down("md"));

  // 按分组统计节点数量
  const groupNodeCounts = useMemo(() => {
    const counts = {};
    allNodes.forEach((node) => {
      const group = node.Group || "未分组";
      counts[group] = (counts[group] || 0) + 1;
    });
    return counts;
  }, [allNodes]);

  // 过滤后的节点列表
  const filteredNodes = useMemo(() => {
    return allNodes.filter((node) => {
      if (nodeGroupFilter !== "all" && node.Group !== nodeGroupFilter) return false;
      if (nodeSourceFilter !== "all" && node.Source !== nodeSourceFilter) return false;
      if (nodeSearchQuery) {
        const query = nodeSearchQuery.toLowerCase();
        if (!node.Name?.toLowerCase().includes(query) && !node.Group?.toLowerCase().includes(query)) {
          return false;
        }
      }
      if (nodeCountryFilter.length > 0) {
        if (!node.LinkCountry || !nodeCountryFilter.includes(node.LinkCountry)) {
          return false;
        }
      }
      return true;
    });
  }, [allNodes, nodeGroupFilter, nodeSourceFilter, nodeSearchQuery, nodeCountryFilter]);

  // 可选节点（排除已选）
  const availableNodes = useMemo(() => {
    return filteredNodes.filter((node) => !formData.selectedNodes.includes(node.Name));
  }, [filteredNodes, formData.selectedNodes]);

  // 已选节点
  const selectedNodesList = useMemo(() => {
    return allNodes.filter((node) => formData.selectedNodes.includes(node.Name));
  }, [allNodes, formData.selectedNodes]);

  return (
    <Dialog open={open} onClose={onClose} maxWidth="lg" fullWidth>
      <DialogTitle>{isEdit ? "编辑订阅" : "添加订阅"}</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ mt: 1 }}>
          <TextField
            fullWidth
            label="订阅名称"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          />

          <Grid container spacing={2}>
            <Grid item xs={6}>
              <FormControl fullWidth>
                <InputLabel>Clash 模板</InputLabel>
                <Select
                  variant={"outlined"}
                  value={formData.clash}
                  label="Clash 模板"
                  onChange={(e) => setFormData({ ...formData, clash: e.target.value })}
                >
                  {templates.map((t) => (
                    <MenuItem key={t.file} value={`./template/${t.file}`}>
                      {t.file}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={6}>
              <FormControl fullWidth>
                <InputLabel>Surge 模板</InputLabel>
                <Select value={formData.surge} label="Surge 模板"
                        onChange={(e) => setFormData({ ...formData, surge: e.target.value })}>
                  {templates.map((t) => (
                    <MenuItem key={t.file} value={`./template/${t.file}`}>
                      {t.file}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
          </Grid>

          <Stack direction="row" spacing={2}>
            <FormControlLabel
              control={<Checkbox checked={formData.udp}
                                 onChange={(e) => setFormData({ ...formData, udp: e.target.checked })} />}
              label="强制开启 UDP"
            />
            <FormControlLabel
              control={<Checkbox checked={formData.cert}
                                 onChange={(e) => setFormData({ ...formData, cert: e.target.checked })} />}
              label="跳过证书验证"
            />
          </Stack>

          <Divider />

          {/* 选择模式 */}
          <Typography variant="subtitle1" fontWeight="bold">
            选择节点
          </Typography>
          <RadioGroup row value={formData.selectionMode}
                      onChange={(e) => setFormData({ ...formData, selectionMode: e.target.value })}>
            <FormControlLabel value="nodes" control={<Radio />} label="手动选择节点" />
            <FormControlLabel value="groups" control={<Radio />} label="动态选择分组" />
            <FormControlLabel value="mixed" control={<Radio />} label="混合模式" />
          </RadioGroup>
          <Typography variant="caption" color="textSecondary">
            {formData.selectionMode === "nodes" && "手动选择具体节点，节点不会随分组变化自动更新"}
            {formData.selectionMode === "groups" && "选择分组，自动包含该分组下的所有节点，节点会随分组变化自动更新"}
            {formData.selectionMode === "mixed" && "同时支持手动选择节点和动态选择分组"}
          </Typography>

          {/* 分组选择 */}
          {(formData.selectionMode === "groups" || formData.selectionMode === "mixed") && (
            <Autocomplete
              multiple
              options={groupOptions}
              value={formData.selectedGroups}
              onChange={(e, newValue) => setFormData({ ...formData, selectedGroups: newValue })}
              renderInput={(params) => <TextField {...params} label="选择分组（动态）" />}
              renderOption={(props, option) => (
                <li {...props}>
                  {option} ({groupNodeCounts[option] || 0} 个节点)
                </li>
              )}
            />
          )}

          {/* 节点选择 */}
          {(formData.selectionMode === "nodes" || formData.selectionMode === "mixed") && (
            <>
              <Grid container spacing={2}>
                <Grid item xs={3}>
                  <FormControl fullWidth size="small">
                    <InputLabel>分组过滤</InputLabel>
                    <Select value={nodeGroupFilter} label="分组过滤"
                            onChange={(e) => setNodeGroupFilter(e.target.value)}>
                      <MenuItem value="all">全部分组 ({allNodes.length})</MenuItem>
                      {groupOptions.map((g) => (
                        <MenuItem key={g} value={g}>
                          {g} ({groupNodeCounts[g] || 0})
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                </Grid>
                <Grid item xs={3}>
                  <FormControl fullWidth size="small">
                    <InputLabel>来源过滤</InputLabel>
                    <Select value={nodeSourceFilter} label="来源过滤"
                            onChange={(e) => setNodeSourceFilter(e.target.value)}>
                      <MenuItem value="all">全部来源</MenuItem>
                      {sourceOptions.map((s) => (
                        <MenuItem key={s} value={s}>
                          {s}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                </Grid>
                <Grid item xs={3}>
                  <Autocomplete
                    multiple
                    size="small"
                    options={countryOptions}
                    value={nodeCountryFilter}
                    onChange={(e, newValue) => setNodeCountryFilter(newValue)}
                    getOptionLabel={(option) => formatCountry(option)}
                    renderInput={(params) => <TextField {...params} label="国家过滤" />}
                    renderOption={(props, option) => <li {...props}>{formatCountry(option)}</li>}
                    limitTags={2}
                  />
                </Grid>
                <Grid item xs={3}>
                  <TextField
                    fullWidth
                    size="small"
                    label="搜索节点"
                    value={nodeSearchQuery}
                    onChange={(e) => setNodeSearchQuery(e.target.value)}
                  />
                </Grid>
              </Grid>

              <NodeTransferBox
                availableNodes={availableNodes}
                selectedNodes={formData.selectedNodes}
                selectedNodesList={selectedNodesList}
                allNodes={allNodes}
                checkedAvailable={checkedAvailable}
                checkedSelected={checkedSelected}
                selectedNodeSearch={selectedNodeSearch}
                onSelectedNodeSearchChange={setSelectedNodeSearch}
                mobileTab={mobileTab}
                onMobileTabChange={setMobileTab}
                matchDownMd={matchDownMd}
                onAddNode={onAddNode}
                onRemoveNode={onRemoveNode}
                onAddAllVisible={onAddAllVisible}
                onRemoveAll={onRemoveAll}
                onToggleAvailable={onToggleAvailable}
                onToggleSelected={onToggleSelected}
                onAddChecked={onAddChecked}
                onRemoveChecked={onRemoveChecked}
                onToggleAllAvailable={onToggleAllAvailable}
                onToggleAllSelected={onToggleAllSelected}
              />
            </>
          )}

          <Divider />

          {/* 延迟和速度过滤 */}
          <Grid container spacing={2}>
            <Grid item xs={6}>
              <TextField
                fullWidth
                label="最大延迟"
                type="number"
                value={formData.DelayTime}
                onChange={(e) => setFormData({ ...formData, DelayTime: Number(e.target.value) })}
                InputProps={{ endAdornment: <InputAdornment position="end">ms</InputAdornment> }}
                helperText="设置筛选节点的延迟阈值，0表示不限制"
              />
            </Grid>
            <Grid item xs={6}>
              <TextField
                fullWidth
                label="最小速度"
                type="number"
                value={formData.MinSpeed}
                onChange={(e) => setFormData({ ...formData, MinSpeed: Number(e.target.value) })}
                InputProps={{ endAdornment: <InputAdornment position="end">MB/s</InputAdornment> }}
                helperText="设置筛选节点的最小下载速度，0表示不限制"
              />
            </Grid>
          </Grid>

          {/* 落地IP国家过滤 */}
          <Grid container spacing={2}>
            <Grid item xs={6}>
              <Autocomplete
                multiple
                options={countryOptions}
                value={formData.CountryWhitelist}
                onChange={(e, newValue) => setFormData({ ...formData, CountryWhitelist: newValue })}
                getOptionLabel={(option) => formatCountry(option)}
                renderInput={(params) => <TextField {...params} label="落地IP国家白名单"
                                                    helperText="只保留这些国家的节点，不选则不限制" />}
                renderOption={(props, option) => <li {...props}>{formatCountry(option)}</li>}
              />
            </Grid>
            <Grid item xs={6}>
              <Autocomplete
                multiple
                options={countryOptions}
                value={formData.CountryBlacklist}
                onChange={(e, newValue) => setFormData({ ...formData, CountryBlacklist: newValue })}
                getOptionLabel={(option) => formatCountry(option)}
                renderInput={(params) => (
                  <TextField {...params} label="落地IP国家黑名单" helperText="排除这些国家的节点（优先级高于白名单）" />
                )}
                renderOption={(props, option) => <li {...props}>{formatCountry(option)}</li>}
              />
            </Grid>
          </Grid>

          {/* 节点名称过滤 */}
          <NodeNameFilter
            whitelistValue={formData.nodeNameWhitelist}
            blacklistValue={formData.nodeNameBlacklist}
            onWhitelistChange={(rules) => setFormData({ ...formData, nodeNameWhitelist: rules })}
            onBlacklistChange={(rules) => setFormData({ ...formData, nodeNameBlacklist: rules })}
          />

          {/* 脚本选择 */}
          <Autocomplete
            multiple
            options={scripts}
            getOptionLabel={(option) => `${option.name} (${option.version})`}
            value={scripts.filter((s) => formData.selectedScripts.includes(s.id))}
            onChange={(e, newValue) => setFormData({ ...formData, selectedScripts: newValue.map((s) => s.id) })}
            renderInput={(params) => (
              <TextField {...params} label="数据处理脚本"
                         helperText="脚本将在查询到节点数据后运行，多个脚本按顺序执行" />
            )}
            renderOption={(props, option) => (
              <li {...props}>
                <Box sx={{ display: "flex", flexDirection: "column" }}>
                  <Typography variant="body1">{option.name}</Typography>
                  <Typography variant="caption" color="textSecondary">
                    版本: {option.version}
                  </Typography>
                </Box>
              </li>
            )}
          />

          <Divider />

          {/* 原名预处理 */}
          <NodeNamePreprocessor
            value={formData.nodeNamePreprocess}
            onChange={(rules) => setFormData({ ...formData, nodeNamePreprocess: rules })}
          />

          {/* 节点命名规则 */}
          <Box>
            <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 2 }}>
              <Typography variant="subtitle1" fontWeight="bold">
                节点命名规则
              </Typography>
              <ButtonGroup size="small" variant="outlined">
                <Tooltip title="可视化构建器 - 拖拽添加变量">
                  <Button
                    onClick={() => setNamingMode("builder")}
                    variant={namingMode === "builder" ? "contained" : "outlined"}
                    startIcon={<BuildIcon />}
                  >
                    {matchDownMd ? "" : "构建器"}
                  </Button>
                </Tooltip>
                <Tooltip title="手动输入模式">
                  <Button
                    onClick={() => setNamingMode("manual")}
                    variant={namingMode === "manual" ? "contained" : "outlined"}
                    startIcon={<EditNoteIcon />}
                  >
                    {matchDownMd ? "" : "手动"}
                  </Button>
                </Tooltip>
              </ButtonGroup>
            </Stack>

            {namingMode === "builder" ? (
              <NodeRenameBuilder value={formData.nodeNameRule}
                                 onChange={(rule) => setFormData({ ...formData, nodeNameRule: rule })} />
            ) : (
              <>
                <TextField
                  fullWidth
                  label="命名规则模板"
                  value={formData.nodeNameRule}
                  onChange={(e) => setFormData({ ...formData, nodeNameRule: e.target.value })}
                  placeholder="例如: [$Protocol]$LinkCountry-$Name"
                  helperText="留空则使用原始名称，仅在访问订阅链接时生效"
                />
                <Box sx={{ mt: 1, p: 1.5, bgcolor: "action.hover", borderRadius: 1 }}>
                  <Typography variant="caption" color="textSecondary" component="div">
                    <strong>可用变量：</strong>
                    <br />• <code>$Name</code> - 系统备注名称 &nbsp;&nbsp; • <code>$LinkName</code> - 原始节点名称
                    <br />• <code>$LinkCountry</code> - 落地IP国家代码 &nbsp;&nbsp; • <code>$Speed</code> - 下载速度
                    <br />• <code>$Delay</code> - 延迟 &nbsp;&nbsp; • <code>$Group</code> - 分组名称
                    <br />• <code>$Source</code> - 来源 &nbsp;&nbsp; • <code>$Index</code> -
                    序号 &nbsp;&nbsp; • <code>$Protocol</code> -
                    协议类型
                  </Typography>
                </Box>
                {formData.nodeNameRule && (
                  <Alert severity="info" sx={{ mt: 1 }}>
                    <Typography variant="body2">
                      <strong>预览：</strong> {previewNodeName(formData.nodeNameRule)}
                    </Typography>
                  </Alert>
                )}
              </>
            )}
          </Box>

          <Divider />

          {/* IP 白名单/黑名单 */}
          <TextField
            fullWidth
            label="IP 黑名单（优先级高于白名单），不允许指定IP访问订阅链接"
            multiline
            rows={2}
            value={formData.IPBlacklist}
            onChange={(e) => setFormData({ ...formData, IPBlacklist: e.target.value })}
            helperText="每行一个 IP 或 CIDR"
          />
          <TextField
            fullWidth
            label="IP 白名单，只允许指定IP访问订阅链接"
            multiline
            rows={2}
            value={formData.IPWhitelist}
            onChange={(e) => setFormData({ ...formData, IPWhitelist: e.target.value })}
            helperText="每行一个 IP 或 CIDR"
          />
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>关闭</Button>
        <Button variant="contained" onClick={onSubmit}>
          确定
        </Button>
      </DialogActions>
    </Dialog>
  );
}
