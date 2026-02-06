<script lang="ts" setup>
import type { AdminRolesApi } from '#/api';

import { onMounted, ref } from 'vue';

import { Card, Table, Tag, message } from 'ant-design-vue';

import { getAdminRoles } from '#/api';

defineOptions({ name: 'AdminRolesList' });

const loading = ref(false);
const tableData = ref<AdminRolesApi.Role[]>([]);

const columns = [
  {
    title: '角色名称',
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: '描述',
    dataIndex: 'description',
    key: 'description',
  },
  {
    title: '权限代码',
    dataIndex: 'perms',
    key: 'perms',
  },
];

async function loadRoles() {
  loading.value = true;
  try {
    const roles = await getAdminRoles();
    tableData.value = roles || [];
  } catch (error: any) {
    const msg =
      error?.response?.data?.message || error?.message || '加载角色列表失败';
    message.error(msg);
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  loadRoles();
});
</script>

<template>
  <div class="p-5">
    <!-- 查询区（角色为只读，无筛选） -->
    <Card class="mb-4">
      <span class="text-gray-500">角色与权限为只读，由后端 perms_json 配置。</span>
    </Card>

    <!-- 表格 -->
    <Card>
      <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      row-key="name"
      :pagination="false"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'perms'">
          <span v-if="!record.perms || !record.perms.length">-</span>
          <template v-else>
            <Tag
              v-for="perm in record.perms"
              :key="perm"
              color="blue"
              class="perm-tag"
            >
              {{ perm }}
            </Tag>
          </template>
        </template>
      </template>
    </Table>
    </Card>
  </div>
</template>

<style scoped>
.perm-tag {
  display: inline-block;
  margin-right: 4px;
  margin-bottom: 4px;
  padding: 2px 6px;
  font-size: 12px;
  background-color: #f0f0f0;
  border-radius: 4px;
}
</style>

