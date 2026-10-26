import InsidePage from "@common/components/aoplatform/InsidePage.tsx";
import { $t } from "@common/locales/index.ts";
import PolicyTabContainer from "./policyTabContainer.tsx";
import DataMasking from "./dataMasking/DataMasking.tsx";
import StrategyPolicy from "./generic/StrategyPolicy";
import { useGlobalContext } from "@common/contexts/GlobalStateContext.tsx";
import { useMemo } from "react";
import { useParams } from "react-router-dom";
import { RouterParams } from "@common/const/type.ts";

const PartitionInsideGlobalPolicy = () => {
  const {state} = useGlobalContext()
  const {serviceId} = useParams<RouterParams>()
  /**
   * tab列表
   */
  const tabItems =useMemo(()=> [
    {
      key: 'dataMasking',
      label: $t('数据脱敏'),
      children: <div className="pr-[40px] preview-document h-full pb-[40px]"><DataMasking publishBtn rowOperation={['edit', 'logs', 'delete']} /></div>
    },
    {
      key: 'limiting',
      label: $t('精细限流'),
      children: <div className="pr-[40px] preview-document h-full pb-[40px]"><StrategyPolicy driver="limiting" publishBtn rowOperation={['edit', 'logs', 'delete']} /></div>
    },
    {
      key: 'fuse',
      label: $t('熔断'),
      children: <div className="pr-[40px] preview-document h-full pb-[40px]"><StrategyPolicy driver="fuse" publishBtn rowOperation={['edit', 'logs', 'delete']} /></div>
    },
    {
      key: 'grey',
      label: $t('灰度发布'),
      children: <div className="pr-[40px] preview-document h-full pb-[40px]"><StrategyPolicy driver="grey" publishBtn rowOperation={['edit', 'logs', 'delete']} /></div>
    },
    {
      key: 'piiMask',
      label: $t('PII 脱敏'),
      children: <div className="pr-[40px] preview-document h-full pb-[40px]"><StrategyPolicy driver="pii-mask" publishBtn rowOperation={['edit', 'logs', 'delete']} /></div>
    },
    {
      key: 'promptInjection',
      label: $t('注入防护'),
      children: <div className="pr-[40px] preview-document h-full pb-[40px]"><StrategyPolicy driver="prompt-injection" publishBtn rowOperation={['edit', 'logs', 'delete']} /></div>
    },
    {
      key: 'contentModeration',
      label: $t('内容审核'),
      children: <div className="pr-[40px] preview-document h-full pb-[40px]"><StrategyPolicy driver="content-moderation" publishBtn rowOperation={['edit', 'logs', 'delete']} /></div>
    }
  ],[state.language])

  return (
      <InsidePage
        pageTitle={$t('全局策略')}
        description={$t("支持对系统全局进行统一的策略配置，从而简化管理并确保一致性。全局策略的优先级比服务策略略低。")}
        showBorder={false}
        scrollPage={false}
      >
        <PolicyTabContainer tabs={tabItems} />
      </InsidePage>
  )
}

export default PartitionInsideGlobalPolicy