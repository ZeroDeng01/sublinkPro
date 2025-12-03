<script setup lang="ts">
import { ref, onMounted, nextTick, computed, watch } from "vue";
import { Search, Refresh } from "@element-plus/icons-vue";
import {
  getNodes,
  AddNodes,
  DelNode,
  UpdateNode,
  getSpeedTestConfig,
  updateSpeedTestConfig,
  runSpeedTest,
} from "@/api/subcription/node";
import {
  getSubSchedulers,
  addSubScheduler,
  updateSubScheduler,
  deleteSubScheduler,
  type SubScheduler,
  type SubSchedulerRequest,
  pullSubScheduler,
} from "@/api/subcription/scheduler";
import { ElMessage, ElMessageBox } from "element-plus";
interface Node {
  ID: number;
  Name: string;
  Link: string;
  DialerProxyName: string;
  CreateDate: string;
  Group: string;
  Speed: number;
  LastCheck: string;
  Source: string;
  DelayTime: number;
}
const tableData = ref<Node[]>([]);
const loading = ref(false);
const Nodelink = ref("");
const NodeOldlink = ref("");
const Nodename = ref("");
const NodeOldname = ref("");
const DialerProxyName = ref("");
const NodeGroup = ref("");
const dialogVisible = ref(false);
const table = ref();
const NodeTitle = ref("");
const radio1 = ref("1");
const groupOptions = ref<string[]>([]);
const sourceOptions = ref<string[]>([]);

// 订阅相关变量
const subSchedulerData = ref<SubScheduler[]>([]);
const subSchedulerDialogVisible = ref(false);
const subSchedulerFormVisible = ref(false);
const subSchedulerForm = ref<SubSchedulerRequest>({
  name: "",
  url: "",
  cron_expr: "",
  enabled: true,
  group: "",
});
const subSchedulerFormTitle = ref("");
const subSchedulerTable = ref();
const subSchedulerSelection = ref<SubScheduler[]>([]);

// Cron表达式验证状态
const cronValidationStatus = ref<{
  isValid: boolean;
  message: string;
}>({
  isValid: true,
  message: "",
});

// 订阅分页
const subCurrentPage = ref(1);
const subPageSize = ref(10);

// 订阅表格数据
const currentSubSchedulerData = computed(() => {
  const start = (subCurrentPage.value - 1) * subPageSize.value;
  const end = start + subPageSize.value;
  return subSchedulerData.value.slice(start, end);
});

async function getnodes() {
  loading.value = true;
  try {
    const { data } = await getNodes();
    tableData.value = data;
    // 提取所有已存在的分组
    const groups = new Set<string>();
    const sources = new Set<string>();
    data.forEach((node: Node) => {
      if (node.Group && node.Group.trim() !== "") {
        groups.add(node.Group);
      }
      if (node.Source && node.Source.trim() !== "") {
        sources.add(node.Source);
      }
    });
    groupOptions.value = Array.from(groups).sort();
    sourceOptions.value = Array.from(sources).sort();
  } catch (error) {
    console.error("获取节点列表失败:", error);
  } finally {
    loading.value = false;
  }
}
onMounted(async () => {
  getnodes();
});
const handleAddNode = () => {
  NodeTitle.value = "添加节点";
  Nodelink.value = "";
  Nodename.value = "";
  radio1.value = "1";
  DialerProxyName.value = "";
  NodeGroup.value = "";
  dialogVisible.value = true;
};

const addnodes = async () => {
  // 分开过滤空行和空格
  let nodelinks = Nodelink.value
    .split(/[\r\n,]/)
    .map((item) => item.trim())
    .filter((item) => item !== "");

  if (NodeTitle.value == "添加节点") {
    // 判断合并还是分开
    if (radio1.value === "1") {
      if (Nodename.value.trim() === "") {
        ElMessage.error("备注不能为空");
        return;
      }
      if (nodelinks.length > 0) {
        const processedLink = nodelinks.join(",");
        await AddNodes({
          link: processedLink,
          name: Nodename.value.trim(),
          dialerProxyName: DialerProxyName.value.trim(),
          group: NodeGroup.value.trim(),
        });
      }
    } else {
      for (let i = 0; i < nodelinks.length; i++) {
        await AddNodes({
          link: nodelinks[i],
          name: "",
          dialerProxyName: DialerProxyName.value.trim(),
          group: NodeGroup.value.trim(),
        });
      }
    }
    ElMessage.success("添加成功");
  } else {
    // 更新节点时处理链接
    const processedLink = nodelinks.join(",");

    await UpdateNode({
      oldname: NodeOldname.value.trim(),
      oldlink: NodeOldlink.value.trim(),
      link: processedLink,
      name: Nodename.value.trim(),
      dialerProxyName: DialerProxyName.value.trim(),
      group: NodeGroup.value.trim(),
    });
    ElMessage.success("更新成功");
  }
  getnodes();
  Nodelink.value = "";
  Nodename.value = "";
  NodeGroup.value = "";
  dialogVisible.value = false;
};

const multipleSelection = ref<Node[]>([]);
const handleSelectionChange = (val: Node[]) => {
  multipleSelection.value = val;
};

// 搜索功能
const searchQuery = ref("");
const searchSourceQuery = ref("");
const groupSearchQuery = ref("");
const searchDelayQuery = ref<number | undefined>(undefined);
const searchSpeedQuery = ref<number | undefined>(undefined);
const handleSearch = () => {
  // filteredTableData 是计算属性，会自动更新
};

const resetSearch = () => {
  searchQuery.value = "";
  searchSourceQuery.value = "";
  groupSearchQuery.value = "";
  searchDelayQuery.value = undefined;
  searchSpeedQuery.value = undefined;
};

// 搜索分组选项
const searchGroupOptions = computed(() => {
  return ["未分组", ...groupOptions.value];
});

// 过滤后的节点列表
const filteredTableData = computed(() => {
  let result = tableData.value;

  // 先按分组过滤
  if (groupSearchQuery.value) {
    if (groupSearchQuery.value === "未分组") {
      result = result.filter((node) => !node.Group || node.Group.trim() === "");
    } else {
      const groupQuery = groupSearchQuery.value.toLowerCase();
      result = result.filter((node) =>
        node.Group.toLowerCase().includes(groupQuery)
      );
    }
  }

  if (searchSourceQuery.value) {
    if (searchSourceQuery.value === "手动添加") {
      result = result.filter(
        (node) => !node.Source || node.Source.trim() === "manual"
      );
    } else {
      const sourceQuery = searchSourceQuery.value.toLowerCase();
      result = result.filter((node) =>
        node.Source.toLowerCase().includes(sourceQuery)
      );
    }
  }

  // 按延迟过滤 (<= 指定值)
  if (searchDelayQuery.value !== undefined && searchDelayQuery.value !== null) {
    result = result.filter(
      (node) => node.DelayTime > 0 && node.DelayTime <= searchDelayQuery.value!
    );
  }

  // 按速度过滤 (> 指定值)
  if (searchSpeedQuery.value !== undefined && searchSpeedQuery.value !== null) {
    result = result.filter((node) => node.Speed > searchSpeedQuery.value!);
  }

  // 再按节点内容过滤
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    result = result.filter(
      (node) =>
        node.Name.toLowerCase().includes(query) ||
        node.Link.toLowerCase().includes(query)
    );
  }

  return result;
});
const selectAll = () => {
  nextTick(() => {
    tableData.value.forEach((row) => {
      table.value.toggleRowSelection(row, true);
    });
  });
};
const handleEdit = (row: any) => {
  radio1.value = "1";
  for (let i = 0; i < tableData.value.length; i++) {
    if (tableData.value[i].ID === row.ID) {
      NodeTitle.value = "编辑节点";
      Nodename.value = tableData.value[i].Name;
      NodeOldname.value = Nodename.value;
      Nodelink.value = tableData.value[i].Link.split(",").join("\n");
      NodeOldlink.value = Nodelink.value;
      DialerProxyName.value = tableData.value[i].DialerProxyName;
      NodeGroup.value = tableData.value[i].Group || "";
      dialogVisible.value = true;
      // value1.value = tableData.value[i].Nodes.map((item) => item.Name)
    }
  }
};
const toggleSelection = () => {
  table.value.clearSelection();
};

const handleDel = (row: any) => {
  ElMessageBox.confirm(`你是否要删除 ${row.Name} ?`, "提示", {
    confirmButtonText: "OK",
    cancelButtonText: "Cancel",
    type: "warning",
  }).then(async () => {
    await DelNode({
      id: row.ID,
    });
    ElMessage({
      type: "success",
      message: "删除成功",
    });
    getnodes();
    // tableData.value = tableData.value.filter((item) => item.ID !== row.ID)
  });
  // console.log('click',row.ID)
};

const selectDel = () => {
  if (multipleSelection.value.length === 0) {
    return;
  }
  ElMessageBox.confirm(
    `你是否要删除选中这些，总计数量${multipleSelection.value.length} ?`,
    "提示",
    {
      confirmButtonText: "OK",
      cancelButtonText: "Cancel",
      type: "warning",
    }
  ).then(() => {
    for (let i = 0; i < multipleSelection.value.length; i++) {
      DelNode({
        id: multipleSelection.value[i].ID,
      });
      tableData.value = tableData.value.filter(
        (item) => item.ID !== multipleSelection.value[i].ID
      );
    }
    ElMessage({
      type: "success",
      message: "删除成功",
    });
  });
};
// 分页显示
const currentPage = ref(1);
const pageSize = ref(10);
const handleSizeChange = (val: number) => {
  pageSize.value = val;
  // console.log(`每页 ${val} 条`);
};

const handleCurrentChange = (val: number) => {
  currentPage.value = val;
  // console.log(`当前页: ${val}`);
};
// 表格数据静态化
const currentTableData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return filteredTableData.value.slice(start, end);
});

// 复制链接
const copyUrl = async (url: string) => {
  const textarea = document.createElement("textarea");
  textarea.value = url;
  document.body.appendChild(textarea);
  textarea.select();
  try {
    const successful = document.execCommand("copy");
    const msg = successful ? "success" : "warning";
    const message = successful ? "复制成功！" : "复制失败！";
    ElMessage({
      type: msg,
      message,
    });
  } catch (err) {
    ElMessage({
      type: "warning",
      message: "复制失败！",
    });
  } finally {
    document.body.removeChild(textarea);
  }
};
const copyInfo = (row: any) => {
  copyUrl(row.Link);
};

// 订阅相关函数
const getSubSchedulerList = async () => {
  try {
    const response = await getSubSchedulers();

    if (response) {
      subSchedulerData.value = response.data || [];
    } else {
      ElMessage.error("获取订阅列表失败");
    }
  } catch (error) {
    console.error("获取订阅列表失败:", error);
    ElMessage.error("获取订阅列表失败");
  }
};

const handleImportSubscription = () => {
  subSchedulerDialogVisible.value = true;
  getSubSchedulerList();
};

const handleAddSubScheduler = () => {
  subSchedulerFormTitle.value = "添加订阅";
  subSchedulerForm.value = {
    name: "",
    url: "",
    cron_expr: "",
    enabled: true,
    group: "",
  };
  subSchedulerFormVisible.value = true;
};

const handleEditSubScheduler = (row: SubScheduler) => {
  subSchedulerFormTitle.value = "编辑订阅";
  subSchedulerForm.value = {
    id: row.ID,
    name: row.Name,
    url: row.URL,
    cron_expr: row.CronExpr,
    enabled: row.Enabled,
    group: row.Group || "",
  };
  subSchedulerFormVisible.value = true;
};

const handlePullSubScheduler = (row: SubScheduler) => {
  ElMessageBox.confirm(`确定要立即更新订阅 "${row.Name}" 吗？`, "确认更新", {
    confirmButtonText: "确定",
    cancelButtonText: "取消",
    type: "warning",
  }).then(async () => {
    try {
      subSchedulerForm.value = {
        id: row.ID,
        name: row.Name,
        url: row.URL,
        cron_expr: row.CronExpr,
        enabled: row.Enabled,
        group: row.Group || "",
      };
      const response = await pullSubScheduler(subSchedulerForm.value);
      if (response) {
        ElMessage.success("提交更新任务成功，请稍后刷新页面查看结果");
        await getSubSchedulerList();
        await getnodes();
      } else {
        ElMessage.error("提交更新任务失败");
      }
    } catch (error) {
      console.error("提交更新任务失败:", error);
      ElMessage.error("提交更新任务失败");
    }
  });
};

const handleDeleteSubScheduler = (row: SubScheduler) => {
  ElMessageBox.confirm(`确定要删除订阅 "${row.Name}" 吗？`, "确认删除", {
    confirmButtonText: "确定",
    cancelButtonText: "取消",
    type: "warning",
  }).then(async () => {
    try {
      const response = await deleteSubScheduler(row.ID);
      if (response) {
        ElMessage.success("删除成功");
        await getSubSchedulerList();
      } else {
        ElMessage.error("删除失败");
      }
    } catch (error) {
      console.error("删除失败:", error);
      ElMessage.error("删除失败");
    }
  });
};

const handleSubSchedulerSelectionChange = (val: SubScheduler[]) => {
  subSchedulerSelection.value = val;
};

const handleBatchDeleteSubScheduler = () => {
  if (subSchedulerSelection.value.length === 0) {
    ElMessage.warning("请选择要删除的项目");
    return;
  }

  ElMessageBox.confirm(
    `确定要删除选中的 ${subSchedulerSelection.value.length} 个订阅吗？`,
    "确认批量删除",
    {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    }
  ).then(async () => {
    try {
      const promises = subSchedulerSelection.value.map((item) =>
        deleteSubScheduler(item.ID)
      );
      await Promise.all(promises);
      ElMessage.success("批量删除成功");
      await getSubSchedulerList();
    } catch (error) {
      console.error("批量删除失败:", error);
      ElMessage.error("批量删除失败");
    }
  });
};

// Cron表达式验证函数
const validateCronExpression = (cron: string): boolean => {
  // 去除首尾空格
  cron = cron.trim();

  // 分割Cron表达式
  const parts = cron.split(/\s+/);

  // 只允许5个部分的Cron表达式
  // 5个部分: 分 时 日 月 周
  if (parts.length !== 5) {
    return false;
  }

  // 验证每个部分的格式
  const ranges = [59, 23, 31, 12, 6]; // 分 时 日 月 周

  for (let i = 0; i < parts.length; i++) {
    const part = parts[i];
    const maxVal = ranges[i];

    // 允许的特殊字符
    if (part === "*" || part === "?") {
      continue;
    }

    // 检查范围表达式 (如: 1-5)
    if (part.includes("-")) {
      const [start, end] = part.split("-").map(Number);
      if (
        isNaN(start) ||
        isNaN(end) ||
        start < 0 ||
        end > maxVal ||
        start > end
      ) {
        return false;
      }
      continue;
    }

    // 检查步长表达式 (如: */5, 0-30/5)
    if (part.includes("/")) {
      const [base, step] = part.split("/");
      if (isNaN(Number(step)) || Number(step) <= 0) {
        return false;
      }

      if (base === "*") {
        continue;
      }

      if (base.includes("-")) {
        const [start, end] = base.split("-").map(Number);
        if (
          isNaN(start) ||
          isNaN(end) ||
          start < 0 ||
          end > maxVal ||
          start > end
        ) {
          return false;
        }
      } else {
        const num = Number(base);
        if (isNaN(num) || num < 0 || num > maxVal) {
          return false;
        }
      }
      continue;
    }

    // 检查列表表达式 (如: 1,3,5)
    if (part.includes(",")) {
      const values = part.split(",").map(Number);
      for (const val of values) {
        if (isNaN(val) || val < 0 || val > maxVal) {
          return false;
        }
      }
      continue;
    }

    // 检查单个数值
    const num = Number(part);
    if (isNaN(num) || num < 0 || num > maxVal) {
      return false;
    }
  }
  return true;
};

// 实时验证Cron表达式
watch(
  () => subSchedulerForm.value.cron_expr,
  (newCron) => {
    if (!newCron || newCron.trim() === "") {
      cronValidationStatus.value = {
        isValid: true,
        message: "",
      };
      return;
    }

    const isValid = validateCronExpression(newCron.trim());
    if (isValid) {
      cronValidationStatus.value = {
        isValid: true,
        message: "Cron表达式格式正确",
      };
    } else {
      // 检查可能的错误原因
      const parts = newCron.trim().split(/\s+/);
      let errorMsg = "Cron表达式格式不正确";

      if (parts.length !== 5) {
        errorMsg = `表达式必须为5个部分，当前有${parts.length}个部分`;
      } else {
        // 如果长度正确但格式错误，检查每个部分
        const partNames = ["分", "时", "日", "月", "周"];
        const ranges = [59, 23, 31, 12, 6]; // 分 时 日 月 周

        for (let i = 0; i < parts.length; i++) {
          const part = parts[i];
          const maxVal = ranges[i];

          // 如果不是通配符，检查是否为有效数值
          if (part !== "*" && part !== "?") {
            if (part.includes("/")) {
              // 步长表达式
              const [base, step] = part.split("/");
              if (isNaN(Number(step)) || Number(step) <= 0) {
                errorMsg = `${partNames[i]}字段步长格式错误：${part}`;
                break;
              }
            } else if (part.includes("-")) {
              // 范围表达式
              const [start, end] = part.split("-").map(Number);
              if (
                isNaN(start) ||
                isNaN(end) ||
                start < 0 ||
                end > maxVal ||
                start > end
              ) {
                errorMsg = `${partNames[i]}字段范围错误：${part}`;
                break;
              }
            } else if (part.includes(",")) {
              // 列表表达式
              const values = part.split(",").map(Number);
              let hasError = false;
              for (const val of values) {
                if (isNaN(val) || val < 0 || val > maxVal) {
                  errorMsg = `${partNames[i]}字段列表值错误：${part}`;
                  hasError = true;
                  break;
                }
              }
              if (hasError) break;
            } else {
              // 单个数值
              const num = Number(part);
              if (isNaN(num) || num < 0 || num > maxVal) {
                errorMsg = `${partNames[i]}字段值超出范围：${part}`;
                break;
              }
            }
          }
        }
      }

      cronValidationStatus.value = {
        isValid: false,
        message: errorMsg,
      };
    }
  }
);

const submitSubSchedulerForm = async () => {
  if (!subSchedulerForm.value.name.trim()) {
    ElMessage.warning("请输入名称");
    return;
  }
  if (!subSchedulerForm.value.url.trim()) {
    ElMessage.warning("请输入URL");
    return;
  }
  if (!subSchedulerForm.value.cron_expr.trim()) {
    ElMessage.warning("请输入Cron表达式");
    return;
  } // 验证Cron表达式格式
  if (!validateCronExpression(subSchedulerForm.value.cron_expr.trim())) {
    ElMessage({
      message: "请输入正确的5字段Cron表达式，格式为：分 时 日 月 周",
      type: "error",
      duration: 5000,
      showClose: true,
    });
    return;
  }

  try {
    let response;
    if (subSchedulerFormTitle.value === "添加订阅") {
      response = await addSubScheduler(subSchedulerForm.value);
      if (response) {
        ElMessage.success("添加成功");
        subSchedulerFormVisible.value = false;
        await getSubSchedulerList();
      } else {
        ElMessage.error("添加失败");
      }
    } else {
      response = await updateSubScheduler(subSchedulerForm.value);
      if (response) {
        ElMessage.success("更新成功");
        subSchedulerFormVisible.value = false;
        await getSubSchedulerList();
      } else {
        ElMessage.error("更新失败");
      }
    }
  } catch (error) {
    console.error("操作失败:", error);
    ElMessage.error("操作失败");
  }
};

const handleSubSizeChange = (val: number) => {
  subPageSize.value = val;
};

const handleSubCurrentChange = (val: number) => {
  subCurrentPage.value = val;
};

// 格式化日期时间
const formatDateTime = (dateTimeString: string) => {
  if (!dateTimeString) return "-";

  try {
    const date = new Date(dateTimeString);
    if (isNaN(date.getTime())) return "-";

    return date.toLocaleString("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  } catch (error) {
    return "-";
  }
};

// 测速配置相关
const speedTestDialogVisible = ref(false);
const speedTestForm = ref({
  cron: "",
  enabled: false,
  mode: "tcp",
  url: "",
  timeout: 5,
  groups: [] as string[],
});

const handleSpeedTestSettings = async () => {
  try {
    const { data } = await getSpeedTestConfig();
    speedTestForm.value = data;
    speedTestDialogVisible.value = true;
  } catch (error) {
    console.error("获取测速配置失败:", error);
    ElMessage.error("获取测速配置失败");
  }
};

const submitSpeedTestSettings = async () => {
  if (speedTestForm.value.enabled && !speedTestForm.value.cron) {
    ElMessage.warning("启用时Cron表达式不能为空");
    return;
  }
  // 简单验证Cron
  if (
    speedTestForm.value.enabled &&
    !validateCronExpression(speedTestForm.value.cron)
  ) {
    ElMessage.error("Cron表达式格式不正确");
    return;
  }

  try {
    await updateSpeedTestConfig(speedTestForm.value);
    ElMessage.success("保存成功");
    speedTestDialogVisible.value = false;
  } catch (error) {
    console.error("保存测速配置失败:", error);
    ElMessage.error("保存测速配置失败");
  }
};

const handleRunSpeedTest = async () => {
  try {
    await runSpeedTest();
    ElMessage.success("测速任务已在后台启动，请稍后刷新查看结果");
  } catch (error) {
    console.error("启动测速任务失败:", error);
    ElMessage.error("启动测速任务失败");
  }
};

const handleSingleSpeedTest = async (row: Node) => {
  try {
    await runSpeedTest([row.ID]);
    ElMessage.success(`节点 ${row.Name} 测速任务已启动`);
  } catch (error) {
    console.error("启动测速任务失败:", error);
    ElMessage.error("启动测速任务失败");
  }
};

const handleBatchSpeedTest = async () => {
  if (multipleSelection.value.length === 0) {
    ElMessage.warning("请选择要测速的节点");
    return;
  }

  const ids = multipleSelection.value.map((node) => node.ID);
  try {
    await runSpeedTest(ids);
    ElMessage.success(`已启动 ${ids.length} 个节点的测速任务`);
  } catch (error) {
    console.error("启动批量测速任务失败:", error);
    ElMessage.error("启动批量测速任务失败");
  }
};

// Cron表达式选项
const cronOptions = [
  { label: "每小时 (0 * * * *)", value: "0 * * * *" },
  { label: "每2小时 (0 */2 * * *)", value: "0 */2 * * *" },
  { label: "每6小时 (0 */6 * * *)", value: "0 */6 * * *" },
  { label: "每12小时 (0 */12 * * *)", value: "0 */12 * * *" },
  { label: "每天0点 (0 0 * * *)", value: "0 0 * * *" },
  { label: "每天3点 (0 3 * * *)", value: "0 3 * * *" },
  { label: "每周一 (0 0 * * 1)", value: "0 0 * * 1" },
];

// 测速URL选项
// 测速URL选项 - 真速度测试 (下载)
const speedTestDownloadOptions = [
  {
    label: "10MB (Cloudflare)",
    value: "https://speed.cloudflare.com/__down?bytes=10000000",
  },
  {
    label: "50MB (Cloudflare)",
    value: "https://speed.cloudflare.com/__down?bytes=50000000",
  },
  {
    label: "100MB (Cloudflare)",
    value: "https://speed.cloudflare.com/__down?bytes=100000000",
  },
];

// 测速URL选项 - 延迟测试 (204)
const speedTest204Options = [
  {
    label: "Cloudflare (cp.cloudflare.com)",
    value: "http://cp.cloudflare.com/generate_204",
  },
  {
    label: "Google (clients3.google.com)",
    value: "http://clients3.google.com/generate_204",
  },
  {
    label: "Google (android.clients.google.com)",
    value: "http://android.clients.google.com/generate_204",
  },
  {
    label: "Gstatic (www.gstatic.com)",
    value: "http://www.gstatic.com/generate_204",
  },
];

// 根据模式动态获取测速URL选项
const currentSpeedTestUrlOptions = computed(() => {
  if (speedTestForm.value.mode === "mihomo") {
    return speedTestDownloadOptions;
  }
  return speedTest204Options;
});

const handleSpeedModeChange = (val: string | number | boolean | undefined) => {
  if (val === "mihomo") {
    speedTestForm.value.url = speedTestDownloadOptions[0].value;
  } else {
    speedTestForm.value.url = speedTest204Options[0].value;
  }
};

</script>

<template>
  <div>
    <el-dialog v-model="dialogVisible" :title="NodeTitle" width="80%">
      <el-input
        v-model="Nodelink"
        placeholder="请输入节点多行使用回车或逗号分开,支持base64格式的url订阅"
        type="textarea"
        style="margin-bottom: 10px"
        :autosize="{ minRows: 2, maxRows: 10 }"
      />
      <el-radio-group
        v-model="radio1"
        class="ml-4"
        v-if="NodeTitle == '添加节点'"
      >
        <el-radio value="1" size="large">合并</el-radio>
        <el-radio value="2" size="large">分开</el-radio>
      </el-radio-group>
      <el-input
        v-model="Nodename"
        placeholder="请输入备注"
        v-if="radio1 != '2'"
      />
      <el-input
        v-model="DialerProxyName"
        placeholder="请输入前置代理节点名称或策略组名称(仅Clash-Meta内核可用)"
      />
      <el-select
        v-model="NodeGroup"
        filterable
        allow-create
        default-first-option
        placeholder="请选择或输入分组名称"
        style="width: 100%"
      >
        <el-option
          v-for="group in groupOptions"
          :key="group"
          :label="group"
          :value="group"
        />
      </el-select>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">关闭</el-button>
          <el-button type="primary" @click="addnodes">确定</el-button>
        </div>
      </template>
    </el-dialog>
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <div class="search-group">
            <el-select
              v-model="groupSearchQuery"
              placeholder="搜索分组"
              style="width: 160px"
              clearable
              filterable
              @change="handleSearch"
            >
              <el-option
                v-for="group in searchGroupOptions"
                :key="group"
                :label="group"
                :value="group"
              />
            </el-select>
            <el-input
              v-model="searchQuery"
              placeholder="搜索节点备注或链接"
              style="width: 200px"
              clearable
              @input="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-select
              v-model="searchSourceQuery"
              placeholder="搜索来源"
              style="width: 160px"
              clearable
              filterable
              @change="handleSearch"
            >
              <el-option label="手动添加" value="手动添加" />
              <el-option
                v-for="source in sourceOptions"
                :key="source"
                :label="source === 'manual' ? '手动添加' : source"
                :value="source"
              />
            </el-select>
            <el-input
              v-model.number="searchDelayQuery"
              placeholder="最大延迟"
              style="width: 200px"
              clearable
              type="number"
              @input="handleSearch"
            >
              <template #append> ms </template>
            </el-input>
            <el-input
              v-model.number="searchSpeedQuery"
              placeholder="最低速度"
              style="width: 200px"
              clearable
              type="number"
              @input="handleSearch"
            >
              <template #append> MB/s </template>
            </el-input>
            <el-button @click="resetSearch">重置</el-button>
          </div>
          <div class="button-group">
            <el-button type="primary" @click="handleAddNode"
              >添加节点</el-button
            >
            <el-button type="warning" @click="handleImportSubscription"
              >导入订阅</el-button
            >
            <el-button type="info" @click="handleSpeedTestSettings"
              >测速设置</el-button
            >
            <el-button type="primary" @click="handleBatchSpeedTest"
              >批量测速</el-button
            >
            <el-button type="success" :icon="Refresh" @click="getnodes"
              >刷新</el-button
            >
          </div>
        </div>
      </template>
      <div style="margin-bottom: 10px"></div>
      <el-table
        ref="table"
        v-loading="loading"
        :data="currentTableData"
        style="width: 100%"
        @selection-change="handleSelectionChange"
        stripe
        :header-cell-style="{ background: '#f5f7fa', color: '#606266' }"
      >
        <el-table-column type="selection" fixed width="55" />
        <el-table-column prop="Name" label="备注" min-width="250">
          <template #default="scope">
            <el-tag type="success" effect="plain">{{ scope.row.Name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="Group" label="分组" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.Group" type="warning" effect="plain">{{
              scope.row.Group
            }}</el-tag>
            <span v-else style="color: #c0c4cc; font-size: 12px">未分组</span>
          </template>
        </el-table-column>
        <el-table-column prop="Source" label="来源" min-width="150">
          <template #default="scope">
            <el-tag
              type="success"
              effect="plain"
              v-if="scope.row.Source === 'manual'"
              >手动添加</el-tag
            >
            <el-tag
              type="warning"
              effect="plain"
              v-if="scope.row.Source !== 'manual'"
              >{{ scope.row.Source }}</el-tag
            >
          </template>
        </el-table-column>
        <el-table-column prop="LinkName" label="节点名称" min-width="200">
          <template #default="scope">
            <el-tag type="success" effect="plain" v-if="scope.row.LinkName">{{
              scope.row.LinkName
            }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column
          prop="Link"
          label="节点链接"
          sortable
          min-width="150"
          :show-overflow-tooltip="true"
        >
          <template #default="scope">
            <span style="font-size: 13px; color: #606266">{{
              scope.row.Link
            }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="LinkAddress" label="LinkAddress" min-width="250">
          <template #default="scope">
            <el-tag
              type="success"
              effect="plain"
              v-if="scope.row.LinkAddress"
              >{{ scope.row.LinkAddress }}</el-tag
            >
          </template>
        </el-table-column>
        <el-table-column prop="LinkHost" label="LinkHost" min-width="150">
          <template #default="scope">
            <el-tag type="success" effect="plain" v-if="scope.row.LinkHost">{{
              scope.row.LinkHost
            }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="LinkPort" label="LinkPort" min-width="150">
          <template #default="scope">
            <el-tag type="success" effect="plain" v-if="scope.row.LinkPort">{{
              scope.row.LinkPort
            }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="DialerProxyName"
          label="前置代理"
          width="120"
          :show-overflow-tooltip="true"
        >
          <template #default="scope">
            <el-tag
              v-if="scope.row.DialerProxyName"
              type="info"
              effect="plain"
              size="small"
              >{{ scope.row.DialerProxyName }}</el-tag
            >
            <span v-else style="color: #c0c4cc; font-size: 12px">-</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="CreateDate"
          label="创建时间"
          width="180"
          sortable
        >
          <template #default="scope">
            <span style="font-size: 13px; color: #909399">{{
              scope.row.CreateDate
            }}</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="DelayTime"
          label="延迟"
          width="120"
          sortable
          fixed="right"
        >
          <template #default="scope">
            <el-tag
              v-if="scope.row.DelayTime > 0"
              :type="
                scope.row.DelayTime < 100
                  ? 'success'
                  : scope.row.DelayTime < 500
                    ? 'warning'
                    : 'danger'
              "
              effect="plain"
            >
              {{ scope.row.DelayTime }}ms
            </el-tag>
            <el-tag
              v-else-if="scope.row.DelayTime === -1"
              type="danger"
              effect="plain"
              >未通过</el-tag
            >
            <span v-else style="color: #c0c4cc">-</span>
            <div
              v-if="scope.row.LastCheck"
              style="font-size: 10px; color: #909399; margin-top: 2px"
            >
              {{ formatDateTime(scope.row.LastCheck).split(" ")[1] }}
            </div>
          </template>
        </el-table-column>
        <el-table-column
          prop="Speed"
          label="速度"
          width="120"
          sortable
          fixed="right"
        >
          <template #default="scope">
            <el-tag v-if="scope.row.Speed > 0" type="success" effect="plain">
              {{ scope.row.Speed.toFixed(2) }}MB/s
            </el-tag>
            <span v-else style="color: #c0c4cc">-</span>
          </template>
        </el-table-column>
        <el-table-column fixed="right" label="操作" width="200">
          <template #default="scope">
            <el-button
              link
              type="success"
              size="small"
              @click="copyInfo(scope.row)"
              >复制</el-button
            >
            <el-button
              link
              type="warning"
              size="small"
              @click="handleSingleSpeedTest(scope.row)"
              >测速</el-button
            >
            <el-button
              link
              type="primary"
              size="small"
              @click="handleEdit(scope.row)"
              >编辑</el-button
            >
            <el-button
              link
              type="danger"
              size="small"
              @click="handleDel(scope.row)"
              >删除</el-button
            >
          </template>
        </el-table-column>
      </el-table>

      <el-empty
        v-if="
          !loading && (!filteredTableData || filteredTableData.length === 0)
        "
        description="暂无节点数据"
      >
        <el-button type="primary" @click="getnodes">
          <el-icon><Refresh /></el-icon>
          重新加载
        </el-button>
      </el-empty>
      <div style="margin-top: 20px"></div>
      <el-button type="info" @click="selectAll()">全选</el-button>
      <el-button type="warning" @click="toggleSelection()">取消选择</el-button>
      <el-button type="danger" @click="selectDel">批量删除</el-button>
      <div style="margin-top: 20px"></div>
      <el-pagination
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        :current-page="currentPage"
        :page-size="pageSize"
        layout="total, sizes, prev, pager, next, jumper"
        :page-sizes="[10, 30, 50, 100]"
        :total="filteredTableData.length"
      />
    </el-card>
    <!-- 导入订阅对话框 -->
    <el-dialog
      v-model="subSchedulerDialogVisible"
      title="订阅管理"
      width="90%"
      :close-on-click-modal="false"
    >
      <div style="margin-bottom: 20px">
        <el-button type="primary" @click="handleAddSubScheduler"
          >添加订阅</el-button
        >
        <el-button type="danger" @click="handleBatchDeleteSubScheduler"
          >批量删除</el-button
        >
        <el-button type="success" :icon="Refresh" @click="getSubSchedulerList"
          >刷新</el-button
        >
      </div>
      <el-table
        ref="subSchedulerTable"
        :data="currentSubSchedulerData"
        style="width: 100%"
        @selection-change="handleSubSchedulerSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="Name" label="名称" min-width="120">
          <template #default="scope">
            <el-tag type="primary">{{ scope.row.Name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="URL"
          label="订阅地址"
          min-width="200"
          :show-overflow-tooltip="true"
        />
        <el-table-column prop="CronExpr" label="Cron表达式" min-width="120" />
        <el-table-column prop="Group" label="分组" min-width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.Group" type="info">{{
              scope.row.Group
            }}</el-tag>
            <span v-else style="color: #909399">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="SuccessCount" label="节点数量" min-width="120" />
        <el-table-column prop="LastRunTime" label="上次运行" min-width="160">
          <template #default="scope">
            <span v-if="scope.row.LastRunTime">
              {{ formatDateTime(scope.row.LastRunTime) }}
            </span>
            <span v-else style="color: #909399">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="NextRunTime" label="下次运行" min-width="160">
          <template #default="scope">
            <span
              v-if="scope.row.NextRunTime"
              :style="{
                color:
                  new Date(scope.row.NextRunTime) <= new Date()
                    ? '#F56C6C'
                    : '#67C23A',
              }"
            >
              {{ formatDateTime(scope.row.NextRunTime) }}
            </span>
            <span v-else style="color: #909399">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="Enabled" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.Enabled ? 'success' : 'danger'">
              {{ scope.row.Enabled ? "启用" : "禁用" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="scope">
            <el-button
              link
              type="success"
              size="small"
              @click="handlePullSubScheduler(scope.row)"
            >
              立即更新
            </el-button>
            <el-button
              link
              type="primary"
              size="small"
              @click="handleEditSubScheduler(scope.row)"
            >
              编辑
            </el-button>
            <el-button
              link
              type="danger"
              size="small"
              @click="handleDeleteSubScheduler(scope.row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top: 20px">
        <el-pagination
          @size-change="handleSubSizeChange"
          @current-change="handleSubCurrentChange"
          :current-page="subCurrentPage"
          :page-size="subPageSize"
          layout="total, sizes, prev, pager, next, jumper"
          :page-sizes="[10, 20, 30, 40]"
          :total="subSchedulerData.length"
        />
      </div>
    </el-dialog>
    <!-- 添加/编辑订阅对话框 -->
    <el-dialog
      v-model="subSchedulerFormVisible"
      :title="subSchedulerFormTitle"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form :model="subSchedulerForm" label-width="120px">
        <el-form-item label="名称" required>
          <el-input
            v-model="subSchedulerForm.name"
            placeholder="请输入订阅名称"
            clearable
          />
        </el-form-item>
        <el-form-item label="订阅地址" required>
          <el-input
            v-model="subSchedulerForm.url"
            placeholder="请输入订阅URL地址（支持clash订阅地址和通用base64订阅地址）"
            clearable
          />
        </el-form-item>
        <el-form-item label="Cron表达式" required>
          <el-select
            v-model="subSchedulerForm.cron_expr"
            filterable
            allow-create
            default-first-option
            placeholder="请选择或输入5字段Cron表达式"
            style="width: 100%"
          >
            <el-option
              v-for="item in cronOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>

          <!-- Cron表达式格式说明 -->
          <div style="margin-top: 5px; font-size: 12px; color: #909399">
            <div><strong>Cron表达式格式 (5字段):</strong> 分 时 日 月 周</div>

            <div style="margin-top: 8px">
              <div
                v-if="
                  subSchedulerForm.cron_expr.trim() &&
                  cronValidationStatus.isValid
                "
                style="margin-bottom: 5px; font-weight: bold; color: #67c23a"
              >
                ✓ {{ cronValidationStatus.message }}
              </div>

              <div
                v-if="
                  subSchedulerForm.cron_expr.trim() &&
                  !cronValidationStatus.isValid
                "
                style="margin-bottom: 5px; font-weight: bold; color: #f56c6c"
              >
                ✗ {{ cronValidationStatus.message }}
              </div>
            </div>

            <div
              v-if="
                subSchedulerForm.cron_expr.trim() &&
                !cronValidationStatus.isValid
              "
              style="
                padding: 8px 12px;
                margin-top: 5px;
                margin-bottom: 10px;
                color: #f56c6c;
                background-color: #fef0f0;
                border-left: 3px solid #f56c6c;
                border-radius: 4px;
              "
            >
              <strong>正确格式示例：</strong> 0 */6 * * * (每6小时执行一次)
            </div>

            <div
              style="
                padding: 8px;
                margin-top: 5px;
                line-height: 1.5;
                background-color: #f5f7fa;
                border-radius: 4px;
              "
            >
              <strong>常用示例:</strong>
              <div>0 */6 * * * - 每6小时执行</div>
              <div>0 0 * * * - 每天0点执行</div>
              <div>0 */2 * * * - 每2小时执行</div>
              <div>0 0 * * 1 - 每周一执行</div>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="分组">
          <el-select
            v-model="subSchedulerForm.group"
            filterable
            allow-create
            default-first-option
            placeholder="请选择或输入分组名称"
            clearable
            style="width: 100%"
          >
            <el-option
              v-for="group in groupOptions"
              :key="group"
              :label="group"
              :value="group"
            />
          </el-select>
          <div style="margin-top: 5px; font-size: 12px; color: #909399">
            设置分组后，从此订阅导入的所有节点将自动归属到此分组
          </div>
        </el-form-item>
        <el-form-item label="启用状态">
          <el-switch
            v-model="subSchedulerForm.enabled"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="subSchedulerFormVisible = false">取消</el-button>
          <el-button type="primary" @click="submitSubSchedulerForm"
            >确定</el-button
          >
        </div>
      </template>
    </el-dialog>
    <!-- 测速配置对话框 -->
    <el-dialog
      v-model="speedTestDialogVisible"
      title="节点测速设置"
      width="500px"
    >
      <el-form :model="speedTestForm" label-width="120px">
        <el-form-item label="启用自动测速">
          <el-switch
            v-model="speedTestForm.enabled"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>
        <el-form-item label="测速模式">
          <el-radio-group
            v-model="speedTestForm.mode"
            @change="handleSpeedModeChange"
          >
            <el-radio label="tcp" value="tcp">仅延迟测试</el-radio>
            <el-radio label="mihomo" value="mihomo">真速度测试</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="测速URL">
          <el-select
            v-model="speedTestForm.url"
            filterable
            allow-create
            default-first-option
            :placeholder="
              speedTestForm.mode === 'mihomo'
                ? '请选择或输入下载测速URL'
                : '请选择或输入204测速URL'
            "
            style="width: 100%"
          >
            <el-option
              v-for="item in currentSpeedTestUrlOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
          <div style="font-size: 12px; color: #909399">
            可以自定义测速URL
            <br />
            <span v-if="speedTestForm.mode === 'mihomo'">
              真速度测试使用可下载资源地址，例如:
              https://speed.cloudflare.com/__down?bytes=10000000
            </span>
            <span v-else>
              延迟测试使用更轻量的204测试地址，例如:
              http://cp.cloudflare.com/generate_204
            </span>
          </div>
        </el-form-item>
        <el-form-item label="超时时间(秒)">
          <el-input-number v-model="speedTestForm.timeout" :min="1" :max="60" />
        </el-form-item>
        <el-form-item label="测速分组">
          <el-select
            v-model="speedTestForm.groups"
            multiple
            filterable
            placeholder="请选择测速分组(留空则测速所有节点)"
            style="width: 100%"
          >
            <el-option
              v-for="group in groupOptions"
              :key="group"
              :label="group"
              :value="group"
            />
          </el-select>
          <div style="font-size: 12px; color: #909399">
            选择要进行测速的节点分组，不选则默认对所有节点进行测速
          </div>
        </el-form-item>
        <el-form-item label="Cron表达式" required>
          <el-select
            v-model="speedTestForm.cron"
            filterable
            allow-create
            default-first-option
            placeholder="请选择或输入Cron表达式"
            style="width: 100%"
          >
            <el-option
              v-for="item in cronOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
          <div style="font-size: 12px; color: #909399">
            格式: 分 时 日 月 周 (例如: 0 */1 * * * 表示每小时执行一次)
          </div>
        </el-form-item>
      </el-form>
      <div style="margin-top: 20px; text-align: center">
        <el-button type="primary" @click="handleRunSpeedTest"
          >立即执行一次测速</el-button
        >
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="speedTestDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitSpeedTestSettings"
            >保存</el-button
          >
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.el-card {
  margin: 10px;
}

.el-input {
  margin-bottom: 10px;
}

.card-header {
  display: flex;
  gap: 15px;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
}

.search-group {
  display: flex;
  gap: 10px;
  align-items: center;
}

.button-group {
  display: flex;
  gap: 10px;
  align-items: center;
}

/* 确保搜索框和按钮高度一致 */
.card-header .el-input {
  margin-bottom: 0;
}

.card-header .el-input :deep(.el-input__wrapper) {
  height: 32px;
}

.card-header .el-button {
  height: 32px;
}

/* 响应式布局优化 */
@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    align-items: stretch;
  }

  .search-group,
  .button-group {
    width: 100%;
    flex-wrap: wrap;
  }

  .search-group .el-input,
  .button-group .el-button {
    flex: 1;
    min-width: 120px;
  }
}
</style>
