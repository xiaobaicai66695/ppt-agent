# Generators 参考文档

本文档是 SlideExecutor 生成 PPT 时的生成器函数参考。

## 导入方式

```python
import sys
from pathlib import Path

script_dir = Path("<工作目录绝对路径>")  # 必须是绝对路径，不要用 __file__ 或 os.getcwd()
generators_pkg_dir = (script_dir / ".." / ".." / "skills" / "visual_designer").resolve()
sys.path.insert(0, str(generators_pkg_dir))

from generators import (
    new_presentation, save_presentation, save_slide,
    generate_title_slide, generate_section_divider, generate_content_slide,
    generate_stat_slide, generate_quote_slide, generate_card_grid,
    generate_timeline, generate_process_flow, generate_two_column,
    generate_three_column, generate_summary_slide, generate_image_text,
    generate_example_detail, generate_deep_dive, generate_agenda,
    generate_case_study, generate_kpi_dashboard,
)
```

## 调用规范（必须严格遵守）

- `new_presentation(palette="xxx")` — 创建空演示文稿（不含幻灯片），每个 PPTX 文件必须单独创建
- 每个 `generate_xxx` 函数必须传入 `prs` 参数（即使为 None，生成器内部会自动 new_presentation）
- `save_slide(slide, output_path)` — 保存单个 slide 为 PPTX 文件（推荐用法）
- **每个 PPTX 文件 = new_presentation + 一次 generate + save_slide**，禁止复用 prs 生成多个文件
- **所有参数都用 keyword 形式传递**（如 `palette="ocean_soft"`），不要依赖位置参数
- **只传函数接受的参数**，所有参数都在各生成器的参数表中列出，未列出的参数请勿传入

## 通用参数

所有 generate 函数共享以下参数：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `prs` | `Optional[Presentation]` | `None` | 已有的 Presentation 对象，为 None 时自动创建 |
| `palette` | `str` | `"ocean_soft"` | 配色方案名，见 palettes.md |
| `source` | `str` | `""` | **数据来源/参考资料**。传入非空字符串时，幻灯片底部渲染灰色小字来源行。格式示例：`"来源: 国家统计局 2025年数据 | https://www.stats.gov.cn"` |

> **强制要求**：使用 search 工具获取数据后，必须在 `source` 参数中列出信息来源 URL 和机构名称。

## 生成器函数参数

### 结构引导类

#### generate_title_slide — 标题页
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"AI大模型技术概述"` |
| subtitle | str | `"从Transformer到GPT-4"` |
| author | str | `"张三"` |
| date | str | `"2025年1月"` |
| kicker | str | `"产品发布 · 2025"` (可选，标题上方小标签) |

#### generate_section_divider — 章节分隔页
| 参数 | 类型 | 示例 |
|------|------|------|
| number | str | `"01"` |
| title | str | `"技术背景"` |
| subtitle | str | `"从感知机到大模型"` |
| kicker | str | `"第三章"` (可选，编号上方小标签) |

#### generate_agenda — 目录页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"目录"` |
| title | str | `"内容概览"` |
| items | `List[str]` | `["01  背景", "02  方法", "03  结论"]` |

#### generate_summary_slide — 总结页
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"总结"` |
| key_points | `List[str]` | `["01  核心结论1", "02  核心结论2"]` (最多4条) |
| thank_you | str | `"感谢聆听"` |
| contact | str | `"{联系方式}"` |
| kicker | str | `"总结"` (可选，标题上方小标签) |

### 内容陈述类

#### generate_content_slide — 普通内容页（兜底类型）
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"深度学习发展历程"` |
| section_header | str | `"{小节标题}"` (可选) |
| bullets | `List[str]` | `["感知机(1957)：首个线性分类器，仅能处理线性可分数据", ...]` (4-6条，每条≤35中文字符) |
| kicker | str | `"要点 · 核心技术"` (可选，标题上方小标签) |

#### generate_quote_slide — 金句/引言页
| 参数 | 类型 | 示例 |
|------|------|------|
| quote | str | `"弱小和无知不是生存的障碍，傲慢才是"` |
| attribution | str | `"— 刘慈欣《三体》"` |
| kicker | str | `"金句"` (可选，引言上方小标签) |

#### generate_image_text — 图文混排页
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"GPT-4多模态能力"` |
| layout | str | `"right-image"` 或 `"left-image"` |
| header | str | `"核心技术突破"` |
| bullets | `List[str]` | `["视觉理解：支持图像输入分析", ...]` (3-4条) |
| kicker | str | `"功能 · 核心"` (可选，标题上方小标签) |
| sub_header | str | `"能力亮点"` (可选，header 与 bullets 之间的次级标题) |

### 对比与并列类

#### generate_two_column — 双栏对比
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"CNN vs Transformer 对比"` |
| left_header | str | `"CNN"` |
| left_bullets | `List[str]` | `["擅长空间特征提取", ...]` (3-5条) |
| right_header | str | `"Transformer"` |
| right_bullets | `List[str]` | `["擅长全局依赖建模", ...]` (3-5条) |
| kicker | str | `"方案对比"` (可选，标题上方小标签) |

#### generate_three_column — 三栏并列
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"三种方案对比"` |
| columns | `List[dict]` | `[{"header": "方案A", "bullets": ["优点1", "优点2"]}, ...]` ×3 |
| kicker | str | `"能力矩阵"` (可选，标题上方小标签) |

#### generate_card_grid — 卡片阵列
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"六大核心能力"` |
| layout | str | `"2x2"` 或 `"2x3"` 或 `"3x2"` |
| cards | `List[dict]` | `[{"header": "智能问答", "body": "基于大模型的NL2SQL"}, ...]` ×4-8 |
| kicker | str | `"能力 · 核心模块"` (可选，标题上方小标签) |
| subtitle | str | `"全方位赋能企业数字化转型"` (可选，标题下方副标题) |

### 流程与关系类

#### generate_timeline — 时间轴
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"AI发展里程碑"` |
| direction | str | `"horizontal"` 或 `"vertical"` |
| nodes | `List[dict]` | `[{"year": "2017", "event": "Transformer论文发表", "icon": "01"}, ...]` ×4-6 |
| kicker | str | `"技术演进"` (可选，标题上方小标签) |
| subtitle | str | `"从深度学习到大模型的时代跨越"` (可选，标题下方副标题) |

#### generate_process_flow — 步骤流程图
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"模型训练流程"` |
| direction | str | `"horizontal"` / `"horizontal_zigzag"` / `"vertical"` |
| steps | `List[dict]` | `[{"num": "01", "title": "数据收集", "desc": "采集多源数据"}, ...]` ×3-6 |
| kicker | str | `"工程实践"` (可选，标题上方小标签) |
| subtitle | str | `"端到端自动化训练流水线"` (可选，标题下方副标题) |

### 数据与指标类

#### generate_stat_slide — 关键数字页
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"系统性能指标"` |
| stats | `List[dict]` | `[{"number": "99.99", "unit": "%", "label": "系统可用性"}, ...]` ×2-4 |
| kicker | str | `"年度成果"` (可选，标题上方小标签) |
| subtitle | str | `"2025财年关键数据一览"` (可选，标题下方副标题) |

#### generate_kpi_dashboard — 指标看板（固定 2x2 布局，最多 4 个 KPI）
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"数据 · 季度增长"` |
| title | str | `"核心业务指标"` |
| kpis | `List[dict]` | `[{"value": "1248K", "label": "月活用户", "delta": "↑38% YoY", "baseline": "去年902K"}, ...]` ×4（固定 2x2 网格，最多 4 个） |
| subtitle | str | `"业务线关键绩效数据"` (可选，标题下方副标题) |

### 内容叙事类（案例/详解）

#### generate_example_detail — 实例详解页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"实例 · 金融风控"` |
| title | str | `"蚂蚁AlphaRisk：实时风控系统"` |
| lede | str | `"日均处理数亿笔交易，风险识别准确率99.99%"` |
| context_block | str | `"金融欺诈每年造成数百亿损失..."` (1-2句背景) |
| solution_block | str | `"基于深度图学习的实时检测..."` (2-3句方案) |
| metrics | `List[dict]` | `[{"value": "99.99%", "label": "准确率", "trend": "↑"}, ...]` ×3 |
| takeaway | str | `"图学习是风控的核心技术方向"` |

#### generate_deep_dive — 深入详解页（双栏）
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"详解 · Transformer架构"` |
| title | str | `"自注意力机制原理"` |
| lede | str | `"一句话概括核心价值"` |
| left_header | str | `"核心要点"` |
| key_points | `List[str]` | `["多头注意力：16个子空间并行建模", ...]` (3-5条) |
| analysis | `List[str]` | `["维度分析1：结论", "维度分析2：结论"]` (2条) |
| right_header | str | `"案例/数据"` |
| case_example | `List[str]` | `["GPT-4：万亿参数，MMLU 86.4%", ...]` (3-4条) |
| data_evidence | `List[str]` | `["推理延迟：320ms→18ms", "训练成本：$63M", ...]` (3条) |
| supplement | `List[str]` | 可选补充信息 (0-2条) |

#### generate_case_study — 案例研究页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"案例 · 智能客服"` |
| title | str | `"某银行AI客服系统"` |
| context | str | `"银行业客服成本高、响应慢..."` (背景) |
| problem | str | `"日均10万+咨询，人工应答率仅60%"` (痛点) |
| solution | str | `"基于RAG+大模型的智能问答..."` (方案) |
| results | `List[dict]` | `[{"metric": "应答率", "value": "95%", "comparison": "提升35%"}, ...]` ×4 |

## 常见错误

| 错误 | 原因 | 修复 |
|------|------|------|
| `save_slide` AttributeError | slide 对象无效（前一步 generate 失败未检查） | 检查 generate 返回值，每个文件独立 new_presentation |
| `No module named 'generators'` | sys.path 指向了 skills/ 而非 skills/visual_designer | 确认路径：`script_dir / ".." / ".." / "skills" / "visual_designer"` |

## 禁止修改的文件

- `skills/visual_designer/generators/__init__.py`
- `skills/visual_designer/generators/base.py`
- `skills/visual_designer/generators/*.py`
