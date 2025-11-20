<script setup lang="ts">
import { ref, onMounted, computed, nextTick } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Search } from "@element-plus/icons-vue";
import {
  getSubs,
  AddSub,
  DelSub,
  UpdateSub,
  SortSub,
} from "@/api/subcription/subs";
import { getTemp } from "@/api/subcription/temp";
import { getNodes } from "@/api/subcription/node";
import QrcodeVue from "qrcode.vue";
import md5 from "md5";
import IPListInput from "@/views/subcription/components/IPListInput.vue";
interface Sub {
  ID: number;
  Name: string;
  CreateDate: string;
  Config: Config;
  Nodes: Node[];
  SubLogs: SubLogs[];
  IPWhitelist: string;
  IPBlacklist: string;
}
interface Node {
  ID: number;
  Name: string;
  Link: string;
  CreateDate: string;
  Sort?: number; // 添加排序字段，可选
  Group?: string; // 分组字段
}
interface Config {
  clash: string;
  surge: string;
  udp: string;
  cert: string;
}
interface SubLogs {
  ID: number;
  IP: string;
  Date: string;
  Addr: string;
  Count: number;
  SubcriptionID: number;
}
interface Temp {
  file: string;
  text: string;
  CreateDate: string;
}
const tableData = ref<Sub[]>([]);
const IPWhitelist = ref("");
const IPBlacklist = ref("");
const Clash = ref("");
const Surge = ref("");
const SubTitle = ref("");
const Subname = ref("");
const oldSubname = ref("");
const dialogVisible = ref(false);
const table = ref();
const NodesList = ref<Node[]>([]);
const value1 = ref<string[]>([]);
const checkList = ref<string[]>([]); // 配置列表
const iplogsdialog = ref(false);
const IplogsList = ref<SubLogs[]>([]);
const qrcode = ref("");
const templist = ref<Temp[]>([]);
const selectedGroup = ref<string>("all"); // 当前选中的分组
const nodeSearchQuery = ref(""); // 节点搜索关键词
async function getsubs() {
  const { data } = await getSubs();
  tableData.value = data;
  processTableData(); // 处理数据，添加父节点ID
}
async function gettemps() {
  const { data } = await getTemp();
  templist.value = data;
  //console.log(templist.value);
}
onMounted(() => {
  getsubs();
  gettemps();
});
onMounted(async () => {
  const { data } = await getNodes();
  NodesList.value = data;
});

const addSubs = async () => {
  const config = JSON.stringify({
    cert: checkList.value.includes("cert"),
    clash: Clash.value.trim(),
    surge: Surge.value.trim(),
    udp: checkList.value.includes("udp"),
  });
  if (SubTitle.value === "添加订阅") {
    await AddSub({
      config: config,
      name: Subname.value.trim(),
      IPWhitelist: IPWhitelist.value,
      IPBlacklist: IPBlacklist.value,
      nodes: value1.value.join(","),
    });
    getsubs();
    ElMessage.success("添加成功");
  } else {
    await UpdateSub({
      config: config,
      name: Subname.value.trim(),
      nodes: value1.value.join(","),
      IPWhitelist: IPWhitelist.value,
      IPBlacklist: IPBlacklist.value,
      oldname: oldSubname.value,
    });
    getsubs();
    ElMessage.success("更新成功");
  }

  dialogVisible.value = false;
};

const multipleSelection = ref<Sub[]>([]);
const handleSelectionChange = (val: Sub[]) => {
  multipleSelection.value = val;
};
const selectAll = () => {
  tableData.value.forEach((row) => {
    table.value.toggleRowSelection(row, true);
  });
};
const handleIplogs = (row: any) => {
  iplogsdialog.value = true;
  nextTick(() => {
    tableData.value.forEach((item) => {
      if (item.ID === row.ID) {
        IplogsList.value = item.SubLogs;
      }
    });
  });
};

// 为树形表格提供唯一的行键，避免子节点与父节点ID冲突，错误的行键会子节点也显示可以展开
const getRowKey = function (row: any): string {
  if (row.Nodes) {
    return row.ID;
  } else {
    return "node_" + row.ID;
  }
};

// 处理数据，为子节点添加父节点ID并设置Sort值，方便排序
const processTableData = () => {
  // 为子节点添加parentId属性
  tableData.value.forEach((subscription) => {
    if (subscription.Nodes) {
      subscription.Nodes.forEach((node, index) => {
        (node as any).parentId = subscription.ID;
        // 如果后端返回了Sort字段，使用后端的值，否则按显示顺序设置
        if (node.Sort === undefined || node.Sort === null) {
          node.Sort = index;
        }
      });

      // 根据Sort字段排序节点
      if (
        subscription.Nodes.length > 0 &&
        subscription.Nodes[0].Sort !== undefined
      ) {
        subscription.Nodes.sort((a, b) => {
          return (a.Sort || 0) - (b.Sort || 0);
        });
      }
    }
  });
};

const toggleSelection = () => {
  table.value.clearSelection();
};

const handleAddSub = () => {
  SubTitle.value = "添加订阅";
  Subname.value = "";
  oldSubname.value = "";
  checkList.value = [];
  Clash.value = "./template/clash.yaml";
  Surge.value = "./template/surge.conf";
  dialogVisible.value = true;
  value1.value = [];
  IPWhitelist.value = "";
  IPBlacklist.value = "";
  selectedGroup.value = "all";
  nodeSearchQuery.value = "";
};
const handleEdit = (row: any) => {
  for (let i = 0; i < tableData.value.length; i++) {
    if (tableData.value[i].ID === row.ID) {
      function toConfig(value: string | Config): Config {
        if (typeof value === "string") {
          return JSON.parse(value) as Config;
        } else {
          return value as Config;
        }
      }
      const config = toConfig(tableData.value[i].Config);
      SubTitle.value = "编辑订阅";
      Subname.value = tableData.value[i].Name;
      oldSubname.value = Subname.value;
      if (config.udp) {
        checkList.value.push("udp");
      }
      if (config.cert) {
        checkList.value.push("cert");
      }
      Clash.value = config.clash;
      Surge.value = config.surge;
      IPWhitelist.value = tableData.value[i].IPWhitelist;
      IPBlacklist.value = tableData.value[i].IPBlacklist;
      dialogVisible.value = true;
      value1.value = tableData.value[i].Nodes.map((item) => item.Name);
      selectedGroup.value = "all";
      nodeSearchQuery.value = "";
    }
  }
};
const handleDel = (row: any) => {
  ElMessageBox.confirm(`你是否要删除 ${row.Name} ?`, "提示", {
    confirmButtonText: "OK",
    cancelButtonText: "Cancel",
    type: "warning",
  }).then(async () => {
    await DelSub({
      id: row.ID,
    });
    getsubs();
    ElMessage({
      type: "success",
      message: "删除成功",
    });
  });
};

const selectDel = () => {
  if (multipleSelection.value.length === 0) {
    return;
  }
  ElMessageBox.confirm(`你是否要删除选中这些 ?`, "提示", {
    confirmButtonText: "OK",
    cancelButtonText: "Cancel",
    type: "warning",
  }).then(() => {
    for (let i = 0; i < multipleSelection.value.length; i++) {
      if (!multipleSelection.value[i].Nodes) {
        continue;
      }
      DelSub({
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
};
// 表格数据静态化
const currentTableData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;

  // 复制表格数据，避免直接修改原始数据
  let data: Sub[] = JSON.parse(JSON.stringify(tableData.value));

  return data.slice(start, end);
});

// 复制链接
const copyUrl = (url: string) => {
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
const handleBase64 = (text: string) => {
  return window.btoa(unescape(encodeURIComponent(text)));
};
const ClientDiaLog = ref(false);
const ClientList = ["v2ray", "clash", "surge"]; // 客户端列表
const ClientUrls = ref<Record<string, string>>({});
const ClientUrl = ref("");
const handleClient = (name: string) => {
  let serverAddress =
    location.protocol +
    "//" +
    location.hostname +
    (location.port ? ":" + location.port : "");
  ClientDiaLog.value = true;
  ClientUrl.value = `${serverAddress}/c/?token=${md5(name)}`;
  ClientList.forEach((item: string) => {
    ClientUrls.value[item] = `${serverAddress}/c/?token=${md5(name)}`;
  });
};

const Qrdialog = ref(false);
const QrTitle = ref("");
const handleQrcode = (url: string, title: string) => {
  Qrdialog.value = true;
  qrcode.value = url;
  QrTitle.value = title;
};
const OpenUrl = (url: string) => {
  window.open(url);
};
const clientradio = ref("1");

// 新增排序相关变量
const sortingSubscriptionId = ref<number | null>(null); // 当前正在排序的订阅ID
const tempNodeSort = ref<{ Name: string; Sort: number }[]>([]); // 临时存储排序数据（使用Name）
const originalNodesOrder = ref<Node[]>([]); // 保存原始顺序，用于取消操作

// 定义拖拽行为所需的变量
const dragSource = ref<number | null>(null);
const dragTarget = ref<number | null>(null);

// 开始拖拽处理
const handleDragStart = (e: DragEvent, nodeId: number) => {
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = "move";
    e.dataTransfer.setData("text/plain", nodeId.toString());
    dragSource.value = nodeId;
  }
};

// 拖拽进入目标区域
const handleDragOver = (e: DragEvent, nodeId: number) => {
  if (e.preventDefault) {
    e.preventDefault();
  }
  if (e.dataTransfer) {
    e.dataTransfer.dropEffect = "move";
  }

  dragTarget.value = nodeId;

  return false;
};

// 拖拽放置
const handleDrop = (
  e: DragEvent,
  targetNodeId: number,
  subscriptionId: number
) => {
  e.stopPropagation();

  // 如果不是在排序模式，或者不是当前被排序的订阅，则忽略
  if (sortingSubscriptionId.value !== subscriptionId) return;

  // 获取被拖动的节点ID
  const sourceNodeId = parseInt(e.dataTransfer?.getData("text/plain") || "0");
  if (sourceNodeId === targetNodeId) return;

  // 在当前排序的订阅中重新排序节点
  const subscription = tableData.value.find((sub) => sub.ID === subscriptionId);
  if (!subscription || !subscription.Nodes) return;

  const sourceIndex = subscription.Nodes.findIndex(
    (node) => node.ID === sourceNodeId
  );
  const targetIndex = subscription.Nodes.findIndex(
    (node) => node.ID === targetNodeId
  );
  if (sourceIndex > -1 && targetIndex > -1) {
    // 移动节点
    const [movedNode] = subscription.Nodes.splice(sourceIndex, 1);
    subscription.Nodes.splice(targetIndex, 0, movedNode);

    // 更新排序字段和临时排序数据
    subscription.Nodes.forEach((node, index) => {
      // 更新节点的Sort属性
      node.Sort = index + 1;

      // 同步更新tempNodeSort中的排序数据（使用Name）
      const sortItem = tempNodeSort.value.find(
        (item) => item.Name === node.Name
      );
      if (sortItem) {
        sortItem.Sort = index + 1;
      } else {
        // 如果不存在则添加
        tempNodeSort.value.push({
          Name: node.Name,
          Sort: index + 1,
        });
      }
    });
  }
  // 重置拖拽状态
  dragSource.value = null;
  dragTarget.value = null;

  return false;
};

// 拖放进入目标元素
const handleDragEnter = (e: DragEvent, nodeId: number) => {
  dragTarget.value = nodeId;
};

// 拖放离开目标元素
const handleDragLeave = () => {
  dragTarget.value = null;
};

// 开始排序
const handleStartSort = (row: any) => {
  sortingSubscriptionId.value = row.ID;
  // 保存原始节点顺序，以便取消时恢复
  originalNodesOrder.value = JSON.parse(JSON.stringify(row.Nodes));

  // 初始化临时排序数据（使用 Name 而不是 ID）
  tempNodeSort.value = row.Nodes.map((node: any, index: number) => ({
    Name: node.Name,
    Sort: node.Sort !== undefined ? node.Sort : index + 1,
  }));

  // 提示用户进入排序模式
  ElMessage({
    type: "info",
    message: "已进入排序模式，可拖动节点进行排序",
    duration: 3000,
  });
};

// 确定排序
const handleConfirmSort = async (row: any) => {
  // 根据当前节点顺序更新Sort值（使用Name）
  row.Nodes.forEach((node: Node, index: number) => {
    const nodeSort = tempNodeSort.value.find((item) => item.Name === node.Name);
    if (nodeSort) {
      nodeSort.Sort = index + 1;
    } else {
      tempNodeSort.value.push({
        Name: node.Name,
        Sort: index + 1,
      });
    }
  });

  // 打印排序结果，格式为后端需要的格式
  var request = {
    ID: row.ID,
    NodeSort: tempNodeSort.value,
  };

  try {
    await SortSub(request);
    ElMessage({
      type: "success",
      message: "节点排序已更新",
      duration: 2000,
    });

    // 重置排序状态
    sortingSubscriptionId.value = null;
    tempNodeSort.value = [];
    originalNodesOrder.value = [];

    // 刷新数据
    await getsubs();
  } catch (error) {
    ElMessage({
      type: "error",
      message: "排序保存失败",
      duration: 2000,
    });
    console.error("排序保存失败:", error);
  }

  // 重置排序状态
  sortingSubscriptionId.value = null;
  tempNodeSort.value = [];
  originalNodesOrder.value = [];
};

// 取消排序
const handleCancelSort = () => {
  // 如果有正在排序的订阅，恢复其节点原始顺序
  if (sortingSubscriptionId.value !== null) {
    for (let i = 0; i < tableData.value.length; i++) {
      if (tableData.value[i].ID === sortingSubscriptionId.value) {
        tableData.value[i].Nodes = JSON.parse(
          JSON.stringify(originalNodesOrder.value)
        );
        break;
      }
    }
  }

  ElMessage({
    type: "info",
    message: "已取消排序操作",
    duration: 2000,
  });

  // 重置排序状态
  sortingSubscriptionId.value = null;
  tempNodeSort.value = [];
  originalNodesOrder.value = [];
};

// 获取所有分组列表
const groupsList = computed(() => {
  const groups = new Set<string>();
  NodesList.value.forEach((node) => {
    if (node.Group && node.Group.trim() !== "") {
      groups.add(node.Group);
    }
  });
  return Array.from(groups).sort();
});

// 根据分组和搜索过滤节点
const filteredNodesList = computed(() => {
  let nodes = NodesList.value;

  // 按分组过滤
  if (selectedGroup.value !== "all") {
    nodes = nodes.filter((node) => node.Group === selectedGroup.value);
  }

  // 按搜索关键词过滤
  if (nodeSearchQuery.value.trim() !== "") {
    const query = nodeSearchQuery.value.toLowerCase();
    nodes = nodes.filter(
      (node) =>
        node.Name.toLowerCase().includes(query) ||
        (node.Group && node.Group.toLowerCase().includes(query))
    );
  }

  return nodes;
});

// 格式化节点显示（带分组信息）
const formatNodeLabel = (node: Node) => {
  if (node.Group && node.Group.trim() !== "") {
    return `${node.Name} [${node.Group}]`;
  }
  return node.Name;
};

// Transfer穿梭框的数据源
const transferData = computed(() => {
  return filteredNodesList.value.map((node) => ({
    key: node.Name,
    label: formatNodeLabel(node),
    disabled: false,
  }));
});

// 按分组统计节点数量
const groupNodeCounts = computed(() => {
  const counts: Record<string, number> = {};
  NodesList.value.forEach((node) => {
    const group = node.Group || "未分组";
    counts[group] = (counts[group] || 0) + 1;
  });
  return counts;
});
</script>

<template>
  <div>
    <el-dialog
      v-model="Qrdialog"
      width="300px"
      style="text-align: center"
      :title="QrTitle"
    >
      <qrcode-vue :value="qrcode" :size="200" level="H" />
      <el-input v-model="qrcode" />
      <el-button @click="copyUrl(qrcode)">复制</el-button>
      <el-button @click="OpenUrl(qrcode)">打开</el-button>
    </el-dialog>

    <el-dialog
      v-model="ClientDiaLog"
      title="客户端(点击二维码获取地址)"
      style="text-align: center"
    >
      <el-row>
        <el-col>
          <el-tag type="success" size="large">自动识别</el-tag>
          <el-button @click="handleQrcode(ClientUrl, '自动识别客户端')"
            >二维码</el-button
          >
        </el-col>
        <el-col
          v-for="(item, index) in ClientUrls"
          :key="index"
          style="margin-bottom: 10px"
        >
          <el-tag type="success" size="large">{{ index }}</el-tag>
          <el-button @click="handleQrcode(`${item}&client=${index}`, index)"
            >二维码</el-button
          >
        </el-col>
      </el-row>
    </el-dialog>

    <el-dialog v-model="iplogsdialog" title="访问记录" width="80%" draggable>
      <template #footer>
        <div class="dialog-footer">
          <el-table :data="IplogsList" border style="width: 100%">
            <el-table-column prop="IP" label="Ip" />
            <el-table-column prop="Count" label="总访问次数" />
            <el-table-column prop="Addr" label="来源" />
            <el-table-column prop="Date" label="最近时间" />
          </el-table>
        </div>
      </template>
    </el-dialog>
    <el-dialog v-model="dialogVisible" :title="SubTitle">
      <el-input v-model="Subname" placeholder="请输入订阅名称" />

      <el-row>
        <el-tag type="primary">clash模版选择</el-tag>
        <el-radio-group v-model="clientradio" class="ml-4">
          <el-radio value="1">本地</el-radio>
          <el-radio value="2">url链接</el-radio>
        </el-radio-group>
        <el-select
          v-model="Clash"
          placeholder="clash模版文件"
          v-if="clientradio === '1'"
        >
          <el-option
            v-for="template in templist"
            :key="template.file"
            :label="template.file"
            :value="'./template/' + template.file"
          />
        </el-select>
        <el-input v-model="Clash" placeholder="clash模版文件" v-else />
      </el-row>
      <el-row>
        <el-tag type="primary">surge模版选择</el-tag>
        <el-radio-group v-model="clientradio" class="ml-4">
          <el-radio value="1">本地</el-radio>
          <el-radio value="2">url链接</el-radio>
        </el-radio-group>
        <el-select
          v-model="Surge"
          placeholder="surge模版文件"
          v-if="clientradio === '1'"
        >
          <el-option
            v-for="template in templist"
            :key="template.file"
            :label="template.file"
            :value="'./template/' + template.file"
          />
        </el-select>
        <el-input v-model="Surge" placeholder="surge模版文件" v-else />
      </el-row>

      <el-row>
        <el-tag type="primary">强制开启选项</el-tag>
        <el-checkbox-group v-model="checkList" style="margin: 5px">
          <el-checkbox :value="'udp'">udp</el-checkbox>
          <el-checkbox :value="'cert'">跳过证书</el-checkbox>
        </el-checkbox-group>
      </el-row>
      <div class="m-4">
        <p style="margin-bottom: 10px; font-weight: 500">选择节点</p>

        <!-- 分组过滤和搜索 -->
        <el-row :gutter="10" style="margin-bottom: 15px">
          <el-col :span="8">
            <el-select
              v-model="selectedGroup"
              placeholder="选择分组"
              style="width: 100%"
              clearable
            >
              <el-option label="全部分组" value="all">
                <span>全部分组</span>
                <span style="float: right; color: #8492a6; font-size: 13px">
                  {{ NodesList.length }}
                </span>
              </el-option>
              <el-option
                v-for="group in groupsList"
                :key="group"
                :label="group"
                :value="group"
              >
                <span>{{ group }}</span>
                <span style="float: right; color: #8492a6; font-size: 13px">
                  {{ groupNodeCounts[group] || 0 }}
                </span>
              </el-option>
            </el-select>
          </el-col>
          <el-col :span="16">
            <el-input
              v-model="nodeSearchQuery"
              placeholder="搜索节点名称或分组"
              clearable
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </el-col>
        </el-row>

        <!-- Transfer 穿梭框 -->
        <el-transfer
          v-model="value1"
          :data="transferData"
          :titles="['可选节点', '已选节点']"
          :button-texts="['移除', '添加']"
          filterable
          :filter-placeholder="'搜索节点'"
          style="text-align: left; display: inline-block"
        >
          <template #default="{ option }">
            <span>{{ option.label }}</span>
          </template>
        </el-transfer>

        <div style="margin-top: 10px; color: #909399; font-size: 12px">
          已选择 <span style="color: #409eff; font-weight: bold">{{ value1.length }}</span> 个节点，
          当前显示 <span style="color: #67c23a; font-weight: bold">{{ filteredNodesList.length }}</span> 个节点
        </div>
      </div>

      <!--IP黑名单，一行一个，支撑CIDR-->
      <IPListInput
        v-model="IPBlacklist"
        title="IP黑名单（优先级高于白名单）"
        class="m-4"
      />

      <!--IP白名单，一行一个，支撑CIDR-->
      <IPListInput v-model="IPWhitelist" title="IP白名单" class="m-4" />

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">关闭</el-button>
          <el-button type="primary" @click="addSubs">确定</el-button>
        </div>
      </template>
    </el-dialog>
    <el-card>
      <el-button type="primary" @click="handleAddSub">添加订阅</el-button>
      <div style="margin-bottom: 10px"></div>
      <el-table
        ref="table"
        :data="currentTableData"
        style="width: 100%"
        stripe
        @selection-change="handleSelectionChange"
        :row-key="getRowKey"
        :tree-props="{ children: 'Nodes' }"
      >
        <el-table-column type="selection" fixed prop="ID" label="id" />
        <el-table-column prop="Name" label="订阅名称 / 节点">
          <template #default="{ row }">
            <!-- 父节点（订阅） -->
            <el-tag v-if="row.Nodes" type="primary">
              {{ row.Name }}
              <span
                v-if="sortingSubscriptionId === row.ID"
                class="sorting-indicator"
              >
                (正在排序)</span
              >
            </el-tag>
            <!-- 子节点（可排序） -->
            <div
              v-else
              :draggable="
                sortingSubscriptionId !== null &&
                row.parentId === sortingSubscriptionId
              "
              @dragstart="(e) => handleDragStart(e, row.ID)"
              @dragover="(e) => handleDragOver(e, row.ID)"
              @drop="(e) => handleDrop(e, row.ID, row.parentId)"
              @dragenter="(e) => handleDragEnter(e, row.ID)"
              @dragleave="handleDragLeave"
              :class="{
                dragging: dragSource === row.ID,
                'drag-over': dragTarget === row.ID,
                'sortable-draggable':
                  sortingSubscriptionId !== null &&
                  row.parentId === sortingSubscriptionId,
              }"
            >
              <el-tag type="success">
                <!--                <template v-if="sortingSubscriptionId !== null && row.parentId === sortingSubscriptionId">-->
                <!--                  <span class="drag-handle">⋮⋮</span>-->
                <!--                </template>-->
                {{ row.Name }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="Link" label="链接" :show-overflow-tooltip="true">
          <template #default="{ row }">
            <div v-if="row.Nodes">
              <el-link
                type="primary"
                size="small"
                @click="handleClient(row.Name)"
                >客户端</el-link
              >
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="CreateDate" label="创建时间" sortable />
        <el-table-column label="操作" width="220">
          <template #default="scope">
            <div v-if="scope.row.Nodes">
              <el-button
                link
                type="primary"
                size="small"
                @click="handleIplogs(scope.row)"
                >记录</el-button
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
                type="primary"
                size="small"
                @click="handleDel(scope.row)"
                >删除</el-button
              >
              <el-button
                v-if="sortingSubscriptionId !== scope.row.ID"
                link
                type="warning"
                size="small"
                @click="handleStartSort(scope.row)"
              >
                排序
              </el-button>
              <el-button
                v-else-if="sortingSubscriptionId === scope.row.ID"
                link
                type="success"
                size="small"
                @click="handleConfirmSort(scope.row)"
              >
                确定修改排序
              </el-button>
              <el-button
                v-if="sortingSubscriptionId === scope.row.ID"
                link
                type="info"
                size="small"
                @click="handleCancelSort()"
              >
                取消
              </el-button>
            </div>
            <div v-else>
              <el-button
                link
                type="primary"
                size="small"
                @click="copyInfo(scope.row)"
                >复制</el-button
              >
            </div>
          </template>
        </el-table-column>
      </el-table>
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
        :page-sizes="[10, 20, 30, 40]"
        :total="tableData.length"
      />
    </el-card>
  </div>
</template>

<style scoped>
.el-card {
  margin: 10px;
}

.el-input {
  margin-bottom: 10px;
}

.el-tag {
  margin: 5px;
}

/* 拖拽相关样式 */
.drag-handle {
  margin-right: 5px;
  font-size: 16px;
  line-height: 1;
  color: #409eff;
  cursor: move;
  user-select: none;
}

.sortable-draggable {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  min-height: 45px;
  padding: 12px 16px;
  margin: 4px 0;
  cursor: move;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border: 2px dashed #d0d7de;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgb(0 0 0 / 10%);
  transition: all 0.3s ease;
}

.sortable-draggable::before {
  position: absolute;
  top: 50%;
  left: 8px;
  width: 4px;
  height: 60%;
  content: "";
  background: linear-gradient(180deg, #409eff 0%, #66b1ff 100%);
  border-radius: 2px;
  opacity: 0.6;
  transition: opacity 0.3s ease;
  transform: translateY(-50%);
}

.sortable-draggable:hover {
  background: linear-gradient(135deg, #e3f2fd 0%, #ecf5ff 100%);
  border-color: #409eff;
  box-shadow: 0 4px 8px rgb(64 158 255 / 25%);
  transform: translateY(-1px);
}

.sortable-draggable:hover::before {
  opacity: 1;
}

/* 被拖拽元素的样式 */
.dragging {
  z-index: 1000;
  color: white;
  background: linear-gradient(45deg, #409eff, #66b1ff) !important;
  border: 2px solid #409eff !important;
  box-shadow: 0 4px 12px rgb(64 158 255 / 40%);
  opacity: 0.8;
  transition: all 0.2s ease-out;
  transform: rotate(2deg) scale(1.02);
}

.dragging .el-tag {
  color: white !important;
  background: rgb(255 255 255 / 20%) !important;
  border: 1px solid rgb(255 255 255 / 30%) !important;
  box-shadow: 0 1px 3px rgb(0 0 0 / 20%);
}

.dragging .drag-handle {
  color: white !important;
}

/* 拖拽目标区域样式 */
.drag-over {
  background: linear-gradient(135deg, #e3f2fd 0%, #bbdefb 100%) !important;
  border: 2px solid #409eff !important;
  box-shadow:
    0 0 0 2px rgb(64 158 255 / 20%),
    inset 0 1px 3px rgb(64 158 255 / 10%);
  transform: scale(1.01);
  animation: dragOverPulse 1s ease-in-out infinite alternate;
}

/* 插入位置指示器 */
.drag-over-before::before {
  position: absolute;
  top: -2px;
  right: 0;
  left: 0;
  height: 4px;
  content: "";
  background: linear-gradient(90deg, #409eff, #66b1ff);
  border-radius: 2px;
  box-shadow: 0 0 8px rgb(64 158 255 / 60%);
  animation: insertIndicator 1s ease-in-out infinite alternate;
}

.drag-over-after::after {
  position: absolute;
  right: 0;
  bottom: -2px;
  left: 0;
  height: 4px;
  content: "";
  background: linear-gradient(90deg, #409eff, #66b1ff);
  border-radius: 2px;
  box-shadow: 0 0 8px rgb(64 158 255 / 60%);
  animation: insertIndicator 1s ease-in-out infinite alternate;
}

/* 动画效果 */
@keyframes dragOverPulse {
  0% {
    background: #e3f2fd;
    transform: scale(1);
  }

  100% {
    background: #bbdefb;
    transform: scale(1.02);
  }
}

@keyframes insertIndicator {
  0% {
    box-shadow: 0 0 4px rgb(64 158 255 / 40%);
    opacity: 0.6;
  }

  100% {
    box-shadow: 0 0 12px rgb(64 158 255 / 80%);
    opacity: 1;
  }
}

.sortable-ghost {
  background: #f0f0f0 !important;
  border: 1px dashed #409eff !important;
  opacity: 0.5;
}

.sortable-chosen {
  background: #e3f2fd !important;
}

.sortable-drag {
  background: #ecf5ff !important;
  opacity: 0.8;
}

/* 排序模式下行样式 */
.el-table__row--sorting {
  background-color: #f8f9fa;
}

.el-table__row--sorting .el-tag {
  position: relative;
}

/* 确保表格单元格的内边距一致 */
.el-table .el-table__cell {
  padding: 8px 0;
}

/* 标签容器统一样式 */
.el-table .el-tag {
  margin: 2px;
  vertical-align: middle;
}

/* 排序指示器样式 */
.sorting-indicator {
  margin-left: 5px;
  color: #409eff;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0% {
    opacity: 0.6;
  }

  50% {
    opacity: 1;
  }

  100% {
    opacity: 0.6;
  }
}

/* Transfer 穿梭框样式优化 */
.el-transfer {
  --el-transfer-panel-width: 300px;
}

.el-transfer-panel {
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgb(0 0 0 / 10%);
}

.el-transfer-panel__header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 8px 8px 0 0;
  font-weight: 500;
}

.el-transfer-panel__header .el-checkbox__label {
  color: white;
}

/* 节点选择区域优化 */
.m-4 {
  margin: 16px 0;
}
</style>
