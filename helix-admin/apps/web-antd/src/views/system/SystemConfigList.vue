<script lang="ts" setup>
import type { SystemConfigApi } from '#/api';

import { onMounted, ref } from 'vue';

import { Button, Card, Form, FormItem, Input, Modal, Table, message } from 'ant-design-vue';

import { getSystemConfigs, updateSystemConfig } from '#/api';

defineOptions({ name: 'SystemConfigList' });

const loading = ref(false);
const tableData = ref<SystemConfigApi.ConfigItem[]>([]);
const editingValues = ref<Record<string, string>>({});
const editingDescriptions = ref<Record<string, string>>({});

const addModalVisible = ref(false);
const addForm = ref({ key: '', value: '', description: '' });
const addSubmitting = ref(false);

function openAddModal() {
  addForm.value = { key: '', value: '', description: '' };
  addModalVisible.value = true;
}

async function handleAddSubmit() {
  const key = addForm.value.key?.trim();
  const value = addForm.value.value?.trim();
  const description = addForm.value.description?.trim() ?? '';
  if (!key) {
    message.error('请输入配置键');
    return;
  }
  if (value === undefined || value === '') {
    message.error('请输入配置值');
    return;
  }
  addSubmitting.value = true;
  try {
    await updateSystemConfig(key, value, description || undefined);
    message.success('新增成功');
    addModalVisible.value = false;
    await loadSystemConfigs();
  } catch (error: any) {
    const msg =
      error?.response?.data?.message || error?.message || '新增失败';
    message.error(msg);
  } finally {
    addSubmitting.value = false;
  }
}

async function loadSystemConfigs() {
  loading.value = true;
  try {
    const configs = await getSystemConfigs();
    tableData.value = configs || [];
    editingValues.value = {};
    editingDescriptions.value = {};
    (configs || []).forEach((item) => {
      editingValues.value[item.key] = item.value ?? '';
      editingDescriptions.value[item.key] = item.description ?? '';
    });
  } catch (error: any) {
    const msg =
      error?.response?.data?.message || error?.message || '加载系统配置失败';
    message.error(msg);
  } finally {
    loading.value = false;
  }
}

async function handleSave(key: string) {
  const value = editingValues.value[key];
  if (value === undefined) {
    message.error('配置值不能为空');
    return;
  }
  const description = editingDescriptions.value[key] ?? '';
  try {
    await updateSystemConfig(key, value, description || undefined);
    message.success('保存成功');
    await loadSystemConfigs();
  } catch (error: any) {
    const msg =
      error?.response?.data?.message || error?.message || '保存失败';
    message.error(msg);
  }
}

onMounted(() => {
  loadSystemConfigs();
});
</script>

<template>
  <div class="p-5">
    <Card class="mb-4">
      <div class="flex items-center justify-between">
        <span class="text-gray-500">
          系统配置为键值对，由后端存储。修改后点击「保存」生效。
        </span>
        <Button type="primary" @click="openAddModal">新增配置</Button>
      </div>
    </Card>

    <Card>
      <Table
        :data-source="tableData"
        :loading="loading"
        :pagination="false"
        row-key="key"
      >
        <Table.Column title="配置键" data-index="key" width="220" />
        <Table.Column title="配置值" width="400">
          <template #default="{ record }">
            <Input
              v-model:value="editingValues[record.key]"
              :placeholder="`请输入 ${record.key} 的值`"
              allow-clear
            />
          </template>
        </Table.Column>
        <Table.Column title="描述" width="200">
          <template #default="{ record }">
            <Input
              v-model:value="editingDescriptions[record.key]"
              placeholder="选填"
              allow-clear
            />
          </template>
        </Table.Column>
        <Table.Column title="操作" width="100" fixed="right">
          <template #default="{ record }">
            <Button type="primary" size="small" @click="handleSave(record.key)">
              保存
            </Button>
          </template>
        </Table.Column>
      </Table>
    </Card>

    <Modal
      v-model:open="addModalVisible"
      title="新增配置"
      :confirm-loading="addSubmitting"
      @ok="handleAddSubmit"
    >
      <Form :model="addForm" layout="vertical">
        <FormItem label="配置键" required>
          <Input
            v-model:value="addForm.key"
            placeholder="例如 feature.xxx"
          />
        </FormItem>
        <FormItem label="配置值" required>
          <Input
            v-model:value="addForm.value"
            placeholder="请输入值"
          />
        </FormItem>
        <FormItem label="描述">
          <Input
            v-model:value="addForm.description"
            placeholder="选填，用于说明该配置用途"
          />
        </FormItem>
      </Form>
    </Modal>
  </div>
</template>
