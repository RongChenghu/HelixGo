export * from './core';

// 导出 getUserInfoApi - 适配 Vben 框架期望的用户信息接口
import { getAdminMe } from './core/auth';
import type { UserInfo } from '@vben/types';

/**
 * 获取用户信息 - 适配 Vben 框架
 * 内部调用 getAdminMe，并转换为框架期望的格式
 */
export async function getUserInfoApi(): Promise<UserInfo> {
  const me = await getAdminMe();
  
  // 转换为 Vben 框架期望的 UserInfo 格式
  return {
    userId: String(me.id),
    username: me.name,
    realName: me.name,
    roles: me.roles || [],
    avatar: '', // 如果没有头像，使用空字符串
    desc: '', // 用户描述
    homePath: '/dashboard', // 默认首页
    token: '', // accessToken 由 accessStore 管理，这里不需要
  } as UserInfo;
}
