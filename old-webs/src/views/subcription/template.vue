<script setup lang="ts">
import { ref, onMounted, nextTick } from "vue";
import { getTemp, AddTemp, UpdateTemp, DelTemp } from "@/api/subcription/temp";
import MonacoEditor from "@/components/MonacoEditor/index.vue";
import { ElMessage, ElMessageBox } from "element-plus";

interface Temp {
  file: string;
  text: string;
  CreateDate: string;
}
const tableData = ref<Temp[]>([]);
const Tempoldname = ref("");
const Tempname = ref("");
const TempText = ref("");
const dialogVisible = ref(false);
const table = ref();
const TempTitle = ref("");
const radio1 = ref("1");

async function gettemps() {
  const { data } = await getTemp();
  tableData.value = data;
}
onMounted(async () => {
  gettemps();
});
const handleAddTemp = () => {
  TempTitle.value = "添加模版";
  Tempname.value = "";
  TempText.value = "";
  radio1.value = "1";
  dialogVisible.value = true;
};
const addtemp = async () => {
  if (TempTitle.value == "添加模版") {
    await AddTemp({
      filename: Tempname.value.trim(),
      text: TempText.value.trim(),
    });
    ElMessage.success("添加成功");
  } else {
    await UpdateTemp({
      filename: Tempname.value.trim(),
      oldname: Tempoldname.value.trim(),
      text: TempText.value.trim(),
    });
    ElMessage.success("更新成功");
  }
  gettemps();
  Tempname.value = "";
  TempText.value = "";
  dialogVisible.value = false;
};

const multipleSelection = ref<Temp[]>([]);
const handleSelectionChange = (val: Temp[]) => {
  multipleSelection.value = val;
};
const selectAll = () => {
  nextTick(() => {
    tableData.value.forEach((row) => {
      table.value.toggleRowSelection(row, true);
    });
  });
};
const handleEdit = (row: any) => {
  for (let i = 0; i < tableData.value.length; i++) {
    if (tableData.value[i].file === row.file) {
      TempTitle.value = "编辑模版";
      Tempname.value = tableData.value[i].file;
      Tempoldname.value = Tempname.value;
      TempText.value = tableData.value[i].text;
      dialogVisible.value = true;
      // value1.value = tableData.value[i].Nodes.map((item) => item.Name)
    }
  }
};
const toggleSelection = () => {
  table.value.clearSelection();
};

const handleDel = (row: any) => {
  ElMessageBox.confirm(`你是否要删除 ${row.file} ?`, "提示", {
    confirmButtonText: "OK",
    cancelButtonText: "Cancel",
    type: "warning",
  }).then(async () => {
    await DelTemp({
      filename: row.file,
      type: row.type,
    });
    ElMessage({
      type: "success",
      message: "删除成功",
    });
    gettemps();
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
      DelTemp({
        filename: multipleSelection.value[i].file,
      });
      tableData.value = tableData.value.filter(
        (item) => item.file !== multipleSelection.value[i].file
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
  return tableData.value.slice(start, end);
});
</script>

<template>
  <div>
    <el-dialog v-model="dialogVisible" :title="TempTitle" width="80%">
      <MonacoEditor
        v-model="TempText"
        language="yaml"
        height="400px"
        style="margin-bottom: 10px"
      />
      <el-input v-model="Tempname" placeholder="模版文件名" />

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">关闭</el-button>
          <el-button type="primary" @click="addtemp">确定</el-button>
        </div>
      </template>
    </el-dialog>
    <el-card>
      <el-button type="primary" @click="handleAddTemp">添加模版</el-button>
      <div style="margin-bottom: 10px"></div>
      <el-table
        ref="table"
        :data="currentTableData"
        :style="{ width: '100%' }"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" fixed prop="ID" label="id" />
        <el-table-column prop="file" label="模版文件名">
          <template #default="scope">
            <el-tag type="success">{{ scope.row.file }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="create_date" label="创建时间" sortable />
        <el-table-column fixed="right" label="操作" width="120">
          <template #default="scope">
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
</style>
@/api/subcription/temp
