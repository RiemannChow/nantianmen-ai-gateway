// 导入echarts核心模块
import * as echarts from 'echarts/core'
// 导入主题JSON
import themeJson from './nantianmen-ai-gateway-chart-palette.json'

// 全局注册主题
export function registerNantianmenAiGatewayTheme() {
  echarts.registerTheme('nantianmen-ai-gateway', themeJson)
}

// 导出主题名称，方便组件使用
export const THEME_NAME = 'nantianmen-ai-gateway'
