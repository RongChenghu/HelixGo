<script lang="ts" setup>
import type { SystemConfigApi } from '#/api';

import { onMounted, ref } from 'vue';

import {
  Button,
  Card,
  Input,
  InputNumber,
  Switch,
  Table,
  message,
  Alert,
} from 'ant-design-vue';
import { useAuthStore } from '#/store';
import { useRouter } from 'vue-router';
import { getSystemConfigs, updateSystemConfig } from '#/api';
import { createAccess } from '#/router/access';

const access = createAccess();

defineOptions({ name: 'SystemConfigList' });

// 表格数据
const loading = ref(false);
const tableData = ref<SystemConfigApi.ConfigItem[]>([]);
const editingValues = ref<Record<string, string>>({});

// 需要显示为开关的配置项
const switchConfigKeys = ['interaction_enabled', 'withdraw_enabled'];

// 需要显示为 JSON 对象表单的配置项
const jsonObjectConfigKeys: string[] = [];

// 判断是否为开关类型的配置
function isSwitchConfig(key: string): boolean {
  return switchConfigKeys.includes(key);
}

// 判断是否为 JSON 对象类型的配置
function isJsonObjectConfig(key: string): boolean {
  return jsonObjectConfigKeys.includes(key);
}

// 解析 JSON 对象配置
function parseJsonObjectConfig(value: string): Record<string, number> {
  try {
    const parsed = JSON.parse(value);
    if (typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)) {
      return parsed;
    }
  } catch (e) {
    // 解析失败，返回默认值
  }
  return { '0': 0.4, '1': 0.45, '2': 0.14, '3': 0.01 };
}

// JSON 对象配置的编辑值
const jsonObjectEditingValues = ref<Record<string, Record<string, number>>>({});

// 初始化 JSON 对象配置的编辑值
function initJsonObjectConfig(key: string, value: string) {
  const parsed = parseJsonObjectConfig(value);
  jsonObjectEditingValues.value[key] = { ...parsed };
}

// 计算概率总和
function calculateSum(key: string): number {
  const values = jsonObjectEditingValues.value[key];
  if (!values) return 0;
  return Object.values(values).reduce((sum, val) => sum + (val || 0), 0);
}

// 更新 JSON 对象配置的值
function updateJsonObjectValue(key: string, slotKey: string, value: number | null | undefined) {
  if (!jsonObjectEditingValues.value[key]) {
    jsonObjectEditingValues.value[key] = {};
  }
  jsonObjectEditingValues.value[key][slotKey] = value ?? 0;
}

// 保存 JSON 对象配置
async function handleSaveJsonObject(key: string) {
  try {
    const values = jsonObjectEditingValues.value[key];
    if (!values) {
      message.error('配置值不能为空');
      return;
    }
    
    // 计算总和
    const sum = calculateSum(key);
    
    // 如果总和为 0，提示错误
    if (sum === 0) {
      message.error('概率总和不能为 0');
      return;
    }
    
    // 归一化：使总和为 1
    const normalized: Record<string, number> = {};
    for (const [slotKey, prob] of Object.entries(values)) {
      normalized[slotKey] = parseFloat((prob / sum).toFixed(4));
    }
    
    // 转换为 JSON 字符串保存
    const jsonString = JSON.stringify(normalized);
    await updateSystemConfig(key, jsonString);
    message.success('保存成功（已自动归一化）');
    // 刷新列表
    await loadSystemConfigs();
  } catch (error: any) {
    console.error('Failed to update system config:', error);
    if (error?.response?.data?.error === 'CONFIG_NOT_FOUND') {
      message.error('配置项不存在');
    } else {
      message.error('保存失败');
    }
  }
}

// 加载系统配置列表
async function loadSystemConfigs() {
  try {
    loading.value = true;
    const configs = await getSystemConfigs();
    tableData.value = configs || [];
    // 初始化编辑值
    editingValues.value = {};
    jsonObjectEditingValues.value = {};
    configs.forEach((item) => {
      editingValues.value[item.key] = item.value;
      // 如果是 JSON 对象配置，初始化解析
      if (isJsonObjectConfig(item.key)) {
        initJsonObjectConfig(item.key, item.value);
      }
    });
  } catch (error) {
    console.error('Failed to load system configs:', error);
    message.error('加载系统配置失败，暂无权限');
  } finally {
    loading.value = false;
  }
}

// 处理开关变化
function handleSwitchChange(key: string, checked: unknown) {
  const boolValue = checked === true || checked === 'true' || checked === 1;
  editingValues.value[key] = boolValue ? 'true' : 'false';
  // 开关变化时自动保存
  handleSave(key);
}

// 保存配置
async function handleSave(key: string) {
  try {
    const value = editingValues.value[key];
    if (value === undefined) {
      message.error('配置值不能为空');
      return;
    }
    
    await updateSystemConfig(key, value);
    message.success('保存成功');
    // 刷新列表
    await loadSystemConfigs();
  } catch (error: any) {
    console.error('Failed to update system config:', error);
    if (error?.response?.data?.error === 'CONFIG_NOT_FOUND') {
      message.error('配置项不存在');
    } else {
      message.error('保存失败');
    }
  }
}
const router = useRouter();
const authStore = useAuthStore();

// ✅ 系统配置白名单（前端兜底）
const SYSTEM_CONFIG_ALLOWLIST_IDS = new Set<number>([2, 4]);
// 初始化加载
onMounted(() => {
  const adminId = Number(authStore.currentUser?.id);

  if (!Number.isFinite(adminId) || !SYSTEM_CONFIG_ALLOWLIST_IDS.has(adminId)) {
    router.replace('/403');
    return;
  }
  if (!access.can('*')) {
    message.error('暂无权限');
    return;
  }
  loadSystemConfigs();
});
</script>

<template>
  <div class="p-5">
    <Card>
      <Table
        :data-source="tableData"
        :loading="loading"
        :pagination="false"
        row-key="key"
      >
        <Table.Column title="配置键" data-index="key" width="200" />
        <Table.Column title="配置值" width="500">
          <template #default="{ record }">
            <!-- 开关类型配置 -->
            <div v-if="isSwitchConfig(record.key)" class="flex items-center gap-2">
              <Switch
                :checked="editingValues[record.key] === 'true'"
                @change="(checked) => handleSwitchChange(record.key, checked)"
              />
              <span :class="editingValues[record.key] === 'true' ? 'text-green-600' : 'text-gray-500'">
                {{ editingValues[record.key] === 'true' ? '开启' : '关闭' }}
              </span>
            </div>
            <!-- JSON 对象配置（命中名额分布） -->
            <div v-else-if="isJsonObjectConfig(record.key)" class="space-y-2">
              <div class="mb-2 text-xs text-gray-500">
                💡 提示：数值越高，中雷机会越大。例如：1人命中设置为0.66，表示66%的概率会抽中1人命中
              </div>
              <div class="grid grid-cols-2 gap-2">
                <div
                  v-for="slotKey in ['0', '1', '2', '3']"
                  :key="slotKey"
                  class="flex items-center gap-2"
                >
                  <span class="text-sm w-16">{{ slotKey }} 人命中:</span>
                  <InputNumber
                    :value="jsonObjectEditingValues[record.key]?.[slotKey]"
                    :min="0"
                    :max="1"
                    :step="0.01"
                    :precision="4"
                    style="width: 100px"
                    @change="(val: unknown) => updateJsonObjectValue(record.key, slotKey, typeof val === 'number' ? val : null)"
                  />
                  <span class="text-xs text-gray-400">概率</span>
                </div>
              </div>
              <Alert
                v-if="calculateSum(record.key) !== 1"
                :message="`概率总和: ${calculateSum(record.key).toFixed(4)} (保存时将自动归一化为 1)`"
                type="info"
                show-icon
                style="margin-top: 8px"
              />
              <Alert
                v-else
                message="概率总和为 1，可以直接保存"
                type="success"
                show-icon
                style="margin-top: 8px"
              />
            </div>
            <!-- hit_rate 特殊提示 -->
            <div v-else-if="record.key === 'hit_rate'" class="space-y-2">
              <Input
                :value="editingValues[record.key]"
                :placeholder="`请输入${record.description || record.key}的值`"
                @update:value="(val) => editingValues[record.key] = val"
              />
              <Alert
                message="重要提示：这是每个红包的中雷概率，不是每局命中人数！"
                description="例如：设置0.5表示每个红包有50%概率中雷。每局5人时，平均约2.5人中雷，但单局可能0-5人都可能。局数越多，数据越接近平均值。"
                type="warning"
                show-icon
                style="margin-top: 8px"
              />
            </div>
            <!-- 普通文本输入 -->
            <Input
              v-else
              :value="editingValues[record.key]"
              :placeholder="`请输入${record.description || record.key}的值`"
              @update:value="(val) => editingValues[record.key] = val"
            />
          </template>
        </Table.Column>
        <Table.Column title="描述" data-index="description" />
        <Table.Column title="操作" width="120" fixed="right">
          <template #default="{ record }">
            <Button
              v-if="isJsonObjectConfig(record.key)"
              type="primary"
              size="small"
              @click="handleSaveJsonObject(record.key)"
            >
              保存
            </Button>
            <Button
              v-else-if="!isSwitchConfig(record.key)"
              type="primary"
              size="small"
              @click="handleSave(record.key)"
            >
              保存
            </Button>
            <span v-else class="text-gray-400 text-sm">开关已自动保存</span>
          </template>
        </Table.Column>
      </Table>
    </Card>
  </div>
</template>
