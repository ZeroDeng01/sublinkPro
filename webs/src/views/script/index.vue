<script setup lang="ts">
import { ref, reactive, onMounted, nextTick, computed } from "vue";
import {
  getScriptList,
  addScript,
  updateScript,
  deleteScript,
} from "@/api/script";
import { ElMessage, ElMessageBox } from "element-plus";
import MonacoEditor from "@/components/MonacoEditor/index.vue";

interface Script {
  id?: number;
  name: string;
  version: string;
  content: string;
  created_at?: string;
  updated_at?: string;
}

const list = ref<Script[]>([]);
const listLoading = ref(true);
const dialogFormVisible = ref(false);
const dialogStatus = ref("");
const textMap = {
  update: "编辑",
  create: "创建",
};
const temp = reactive<Script>({
  id: undefined,
  name: "",
  version: "0.0.0",
  content: "",
});
const rules = {
  name: [{ required: true, message: "脚本名称不能为空", trigger: "blur" }],
  content: [{ required: true, message: "脚本代码不能为空", trigger: "blur" }],
  version: [{ required: true, message: "版本不能为空", trigger: "blur" }],
};

const dataForm = ref();
const table = ref();
const multipleSelection = ref<Script[]>([]);

const searchQuery = reactive({
  name: "",
  version: "",
});

const formatDate = (dateStr: string | undefined) => {
  if (!dateStr) return "";
  const date = new Date(dateStr);
  return date.toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  });
};

// Pagination
const currentPage = ref(1);
const pageSize = ref(10);

const getList = () => {
  listLoading.value = true;
  getScriptList().then((response) => {
    list.value = response.data;
    listLoading.value = false;
  });
};

const resetTemp = () => {
  Object.assign(temp, {
    id: undefined,
    name: "",
    version: "0.0.0",
    content:
      "//修改节点列表\n" +
      "function filterNode(nodes, clientType) {\n" +
      "    // nodes: 节点列表\n" +
      "    // 数据结构如下\n" +
      "    // [\n" +
      "    //     {\n" +
      '    //         "ID": 1,\n' +
      '    //         "Link": "vmess://4564564646",\n' +
      '    //         "Name": "xx订阅_US-CDN-SSL",\n' +
      '    //         "LinkName": "US-CDN-SSL",\n' +
      '    //         "LinkAddress": "xxxxxxxxx.net:443",\n' +
      '    //         "LinkHost": "xxxxxxxxx.net",\n' +
      '    //         "LinkPort": "443",\n' +
      '    //         "DialerProxyName": "",\n' +
      '    //         "CreateDate": "",\n' +
      '    //         "Source": "manual",\n' +
      '    //         "SourceID": 0,\n' +
      '    //         "Group": "自用",\n' +
      '    //         "Speed": 110,\n' +
      '    //         "LastCheck": "2025-11-26 23:49:58"\n' +
      "    //     }, {\n" +
      '    //         "ID": 2,\n' +
      '    //         "Link": "vmess://456456464611111",\n' +
      '    //         "Name": "xx订阅_US-CDN-SSL1",\n' +
      '    //         "LinkName": "US-CDN-SSL1",\n' +
      '    //         "LinkAddress": "xxxxxxxxx1.net:443",\n' +
      '    //         "LinkHost": "xxxxxxxxx1.net",\n' +
      '    //         "LinkPort": "443",\n' +
      '    //         "DialerProxyName": "",\n' +
      '    //         "CreateDate": "",\n' +
      '    //         "Source": "manual",\n' +
      '    //         "SourceID": 0,\n' +
      '    //         "Group": "自用",\n' +
      '    //         "Speed": 100,\n' +
      '    //         "LastCheck": "2025-11-26 23:49:58"\n' +
      "    //     }\n" +
      "    // ]\n" +
      "    // clientType: 客户端类型\n" +
      "    // 返回值: 修改后节点列表\n" +
      "    return nodes;\n" +
      "}\n" +
      "//修改订阅文件\n" +
      "function subMod( input, clientType) {\n" +
      "    // input: 原始输入内容,不同客户端订阅文件也不一样\n" +
      "    // clientType: 客户端类型\n" +
      "    // 返回值: 修改后的内容字符串\n" +
      "    return input; // 注意：此处示例仅为示意，实际应返回处理后的字符串\n" +
      "}",
  });
};

const handleCreate = () => {
  resetTemp();
  dialogStatus.value = "create";
  dialogFormVisible.value = true;
  nextTick(() => {
    dataForm.value?.clearValidate();
  });
};

const createData = () => {
  dataForm.value?.validate((valid: boolean) => {
    if (valid) {
      addScript(temp).then(() => {
        dialogFormVisible.value = false;
        ElMessage.success("创建成功");
        getList();
      });
    }
  });
};

const handleUpdate = (row: Script) => {
  Object.assign(temp, row);
  dialogStatus.value = "update";
  dialogFormVisible.value = true;
  nextTick(() => {
    dataForm.value?.clearValidate();
  });
};

const updateData = () => {
  dataForm.value?.validate((valid: boolean) => {
    if (valid) {
      updateScript(temp).then(() => {
        dialogFormVisible.value = false;
        ElMessage.success("更新成功");
        getList();
      });
    }
  });
};

const handleDelete = (row: Script) => {
  ElMessageBox.confirm("确认删除该脚本吗?", "提示", {
    confirmButtonText: "确定",
    cancelButtonText: "取消",
    type: "warning",
  }).then(() => {
    deleteScript(row).then(() => {
      ElMessage.success("删除成功");
      getList();
    });
  });
};

const handleSelectionChange = (val: Script[]) => {
  multipleSelection.value = val;
};

const selectAll = () => {
  nextTick(() => {
    list.value.forEach((row) => {
      table.value.toggleRowSelection(row, true);
    });
  });
};

const toggleSelection = () => {
  table.value.clearSelection();
};

const selectDel = () => {
  if (multipleSelection.value.length === 0) {
    return;
  }
  ElMessageBox.confirm(`你是否要删除选中这些 ?`, "提示", {
    confirmButtonText: "OK",
    cancelButtonText: "Cancel",
    type: "warning",
  }).then(async () => {
    for (let i = 0; i < multipleSelection.value.length; i++) {
      await deleteScript(multipleSelection.value[i]);
    }
    ElMessage.success("删除成功");
    getList();
  });
};

const handleSizeChange = (val: number) => {
  pageSize.value = val;
};

const handleCurrentChange = (val: number) => {
  currentPage.value = val;
};

const currentTableData = computed(() => {
  let filteredList = list.value;

  if (searchQuery.name) {
    filteredList = filteredList.filter((item) =>
      item.name.toLowerCase().includes(searchQuery.name.toLowerCase())
    );
  }

  if (searchQuery.version) {
    filteredList = filteredList.filter((item) =>
      item.version.toLowerCase().includes(searchQuery.version.toLowerCase())
    );
  }

  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return filteredList.slice(start, end);
});

onMounted(() => {
  getList();
});
</script>

<template>
  <div>
    <el-dialog
      :title="textMap[dialogStatus as keyof typeof textMap]"
      v-model="dialogFormVisible"
      width="80%"
    >
      <el-form
        ref="dataForm"
        :rules="rules"
        :model="temp"
        label-position="left"
        label-width="100px"
        style="margin-left: 50px; margin-right: 50px"
      >
        <el-form-item label="脚本名称" prop="name">
          <el-input v-model="temp.name" />
        </el-form-item>
        <el-form-item label="版本" prop="version">
          <el-input v-model="temp.version" placeholder="默认 0.0.0" />
        </el-form-item>
        <el-form-item label="脚本代码" prop="content">
          <MonacoEditor
            v-model="temp.content"
            language="javascript"
            height="400px"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogFormVisible = false"> 取消 </el-button>
          <el-button
            type="primary"
            @click="dialogStatus === 'create' ? createData() : updateData()"
          >
            确认
          </el-button>
        </div>
      </template>
    </el-dialog>

    <el-card>
      <div class="filter-container" style="margin-bottom: 20px">
        <el-input
          v-model="searchQuery.name"
          placeholder="脚本名称"
          style="width: 200px; margin-right: 10px"
          class="filter-item"
          clearable
        />
        <el-input
          v-model="searchQuery.version"
          placeholder="版本"
          style="width: 200px; margin-right: 10px"
          class="filter-item"
          clearable
        />
        <el-button type="primary" @click="handleCreate"> 添加脚本 </el-button>
      </div>

      <el-table
        ref="table"
        v-loading="listLoading"
        :data="currentTableData"
        :style="{ width: '100%' }"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column label="脚本名称" align="center">
          <template #default="{ row }">
            <el-tag type="success">{{ row.name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="版本" align="center">
          <template #default="{ row }">
            <span>{{ row.version }}</span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" align="center">
          <template #default="{ row }">
            <span>{{ formatDate(row.created_at) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" align="center">
          <template #default="{ row }">
            <span>{{ formatDate(row.updated_at) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="操作"
          align="center"
          width="230"
          class-name="small-padding fixed-width"
        >
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              size="small"
              @click="handleUpdate(row)"
            >
              编辑
            </el-button>
            <el-button
              link
              size="small"
              type="primary"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
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
        :total="list.length"
      />
    </el-card>
  </div>
</template>

<style scoped>
.el-card {
  margin: 10px;
}
</style>
