name: Bug 报告
description: 报告一个缺陷以帮助我们改进
title: "[Bug] "
labels: ["bug"]
body:
  - type: textarea
    id: description
    attributes:
      label: 问题描述
      description: 清晰简洁地描述遇到的问题
    validations:
      required: true
  - type: textarea
    id: repro
    attributes:
      label: 复现步骤
      description: 如何复现该问题
      placeholder: |
        1. 进入 '...'
        2. 调用 '...'
        3. 出现错误
    validations:
      required: true
  - type: textarea
    id: expected
    attributes:
      label: 期望行为
    validations:
      required: true
  - type: input
    id: version
    attributes:
      label: 版本
      description: nantianmen-ai-gateway 版本或 commit
    validations:
      required: true
  - type: dropdown
    id: deploy
    attributes:
      label: 部署方式
      options:
        - Docker Compose
        - 二进制
        - 源码构建
        - Helm/K8s
    validations:
      required: true
  - type: textarea
    id: logs
    attributes:
      label: 相关日志
      description: 请粘贴关键日志（注意脱敏，不要包含密钥）
      render: shell
