<script lang="ts" setup>
import type { AdminAuditApi } from '#/api';

import { ref } from 'vue';

import {
  Button,
  Card,
  DatePicker,
  Form,
  Input,
  Select,
  SelectOption,
  Space,
  Table,
  Tag,
} from 'ant-design-vue';

import { getAuditLogs } from '#/api';

import dayjs, { type Dayjs } from 'dayjs';

defineOptions({ name: 'AuditLogsList' });

// 查询表单
const searchForm = ref({
  keyword: '',
  action: undefined as string | undefined,
  dateRange: undefined as [Dayjs, Dayjs] | undefined,
});

// 表格数据
const loading = ref(false);
const tableData = ref<AdminAuditApi.AuditLog[]>([]);
const total = ref(0);
const pagination = ref({
  current: 1,
  pageSize: 10,
});

// 操作类型选项
const actionOptions = [
  { label: '创建管理员', value: 'admin.create' },
  { label: '启用/禁用', value: 'admin.enable' },
  { label: '重置密码', value: 'admin.reset-password' },
  { label: '设置角色', value: 'admin.set-roles' },
];

// 表格列
const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'id',
    width: 80,
  },
  {
    title: '操作人',
    dataIndex: 'adminUsername',
    key: 'adminUsername',
    width: 120,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 150,
    customRender: ({ record }: { record: AdminAuditApi.AuditLog }) => {
      const actionMap: Record<string, string> = {
        'admin.create': '创建管理员',
        'admin.enable': '启用/禁用',
        'admin.reset-password': '重置密码',
        'admin.set-roles': '设置角色',
      };
      return actionMap[record.action] || record.action;
    },
  },
  {
    title: '目标',
    key: 'target',
    width: 150,
    customRender: ({ record }: { record: AdminAuditApi.AuditLog }) => {
      if (record.targetType && record.targetId) {
        return `${record.targetType}: ${record.targetId}`;
      }
      return '-';
    },
  },
  {
    title: '详情',
    dataIndex: 'payload',
    key: 'payload',
    customRender: ({ record }: { record: AdminAuditApi.AuditLog }) => {
      if (record.payload && typeof record.payload === 'object') {
        const keys = Object.keys(record.payload);
        if (keys.length > 0) {
          return keys.map((key) => `${key}: ${JSON.stringify(record.payload[key])}`).join(', ');
        }
      }
      return '-';
    },
  },
  {
    title: 'IP',
    dataIndex: 'ip',
    key: 'ip',
    width: 120,
  },
  {
    title: '时间',
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 180,
    customRender: ({ record }: { record: AdminAuditApi.AuditLog }) => {
      return dayjs(record.createdAt).format('YYYY-MM-DD HH:mm:ss');
    },
  },
];

// 加载列表
async function loadAuditLogs() {
  loading.value = true;
  try {
    const params: AdminAuditApi.ListParams = {
      keyword: searchForm.value.keyword || undefined,
      action: searchForm.value.action || undefined,
      page: pagination.value.current,
      pageSize: pagination.value.pageSize,
    };

    // 处理日期范围
    if (searchForm.value.dateRange && searchForm.value.dateRange.length === 2) {
      params.from = searchForm.value.dateRange[0].format('YYYY-MM-DD');
      params.to = searchForm.value.dateRange[1].format('YYYY-MM-DD');
    }

    const res = await getAuditLogs(params);
    tableData.value = res.list || [];
    total.value = res.total || 0;
  } catch (error) {
    console.error('Failed to load audit logs:', error);
  } finally {
    loading.value = false;
  }
}

// 搜索
function handleSearch() {
  pagination.value.current = 1;
  loadAuditLogs();
}

// 重置
function handleReset() {
  searchForm.value = {
    keyword: '',
    action: undefined,
    dateRange: undefined,
  };
  pagination.value.current = 1;
  loadAuditLogs();
}

// 分页变化
function handleTableChange(pag: any) {
  pagination.value.current = pag.current ?? 1;
  pagination.value.pageSize = pag.pageSize ?? 10;
  loadAuditLogs();
}

// 初始化
loadAuditLogs();
</script>

<template>
  <div class="p-5">
    <!-- 查询表单 -->
    <Card class="mb-4">
      <Form :model="searchForm" layout="inline">
        <Form.Item label="关键词">
          <Input
            v-model:value="searchForm.keyword"
            placeholder="操作人用户名"
            style="width: 200px"
            @press-enter="handleSearch"
          />
        </Form.Item>
        <Form.Item label="操作类型">
          <Select
            v-model:value="searchForm.action"
            placeholder="全部"
            style="width: 150px"
            allow-clear
          >
            <SelectOption
              v-for="option in actionOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ option.label }}
            </SelectOption>
          </Select>
        </Form.Item>
        <Form.Item label="日期范围">
          <DatePicker.RangePicker
            v-model:value="searchForm.dateRange"
            format="YYYY-MM-DD"
          />
        </Form.Item>
        <Form.Item>
          <Space>
            <Button type="primary" @click="handleSearch">搜索</Button>
            <Button @click="handleReset">重置</Button>
          </Space>
        </Form.Item>
      </Form>
    </Card>

    <!-- 表格 -->
    <Card>
      <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="{
        current: pagination.current,
        pageSize: pagination.pageSize,
        total: total,
        showTotal: (total) => `共 ${total} 条`,
      }"
      @change="handleTableChange"
      />
    </Card>
  </div>
</template>

