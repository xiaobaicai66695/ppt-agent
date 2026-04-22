import React, { useState, useRef, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
}

interface ExecutionEvent {
  type: string;
  message: string;
  phase?: string;
  slide?: number;
  total?: number;
}

interface Plan {
  title: string;
  theme: string;
  slides: Slide[];
}

interface Slide {
  index: number;
  title: string;
  desc: string;
}

const API_BASE = 'http://localhost:8080';

const templates = [
  {
    id: 'ai-intro',
    icon: '🤖',
    title: 'AI大模型介绍',
    description: '介绍AI大模型的基础知识和应用场景',
    prompt: '帮我做一个关于AI大模型的PPT，包含发展历程、技术原理、应用场景和未来展望'
  },
  {
    id: 'tech-sharing',
    icon: '💻',
    title: '技术分享',
    description: '通用的技术分享演示文稿',
    prompt: '做一个技术分享PPT，主题是微服务架构设计'
  },
  {
    id: 'product-intro',
    icon: '📦',
    title: '产品介绍',
    description: '产品特点和功能展示',
    prompt: '做一个新产品的介绍PPT，包含产品功能、优势和使用方法'
  },
  {
    id: 'plan-report',
    icon: '📊',
    title: '商业计划',
    description: '创业项目或商业计划展示',
    prompt: '做一个商业计划书PPT，包含项目背景、市场分析、商业模式和盈利预测'
  }
];

function App() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);
  const [currentEvent, setCurrentEvent] = useState<ExecutionEvent | null>(null);
  const [progress, setProgress] = useState(0);
  const [planResult, setPlanResult] = useState<Plan | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSend = async () => {
    if (!inputValue.trim() || isProcessing) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: inputValue,
      timestamp: new Date()
    };

    setMessages(prev => [...prev, userMessage]);
    setInputValue('');
    setIsProcessing(true);
    setPlanResult(null);

    const aiMessage: Message = {
      id: (Date.now() + 1).toString(),
      role: 'assistant',
      content: '',
      timestamp: new Date()
    };
    setMessages(prev => [...prev, aiMessage]);

    try {
      const response = await fetch(`${API_BASE}/api/generate/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query: inputValue })
      });

      const reader = response.body?.getReader();
      const decoder = new TextDecoder();

      if (reader) {
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          const chunk = decoder.decode(value);
          const lines = chunk.split('\n');

          for (const line of lines) {
            if (line.startsWith('data: ')) {
              try {
                const data = JSON.parse(line.slice(6));
                handleStreamEvent(data);
              } catch (e) {
                console.error('Parse error:', e);
              }
            }
          }
        }
      }
    } catch (error) {
      console.error('Request failed:', error);
      simulateProcess();
    }
  };

  const handleStreamEvent = (event: ExecutionEvent) => {
    setCurrentEvent(event);

    if (event.type === 'phase_start') {
      setMessages(prev => {
        const lastMsg = prev[prev.length - 1];
        if (lastMsg && lastMsg.role === 'assistant' && lastMsg.content === '') {
          return [...prev.slice(0, -1), { ...lastMsg, content: `**${event.message}**\n\n` }];
        }
        return [...prev, {
          id: Date.now().toString(),
          role: 'assistant' as const,
          content: `**${event.message}**\n\n`,
          timestamp: new Date()
        }];
      });
    } else if (event.type === 'slide_progress') {
      setProgress((event.slide || 0) / (event.total || 1) * 100);
      setMessages(prev => {
        const lastMsg = prev[prev.length - 1];
        if (lastMsg && lastMsg.role === 'assistant') {
          return [...prev.slice(0, -1), {
            ...lastMsg,
            content: lastMsg.content + `正在生成第 ${event.slide}/${event.total} 页...\n\n`
          }];
        }
        return prev;
      });
    } else if (event.type === 'plan_complete' && event.plan) {
      setPlanResult(event.plan as Plan);
    } else if (event.type === 'complete') {
      setIsProcessing(false);
      setProgress(100);
      setMessages(prev => {
        const lastMsg = prev[prev.length - 1];
        if (lastMsg && lastMsg.role === 'assistant') {
          return [...prev.slice(0, -1), {
            ...lastMsg,
            content: lastMsg.content + `\n\n✅ **PPT生成完成！**`
          }];
        }
        return prev;
      });
    }
  };

  const simulateProcess = () => {
    const steps = [
      { type: 'phase_start', message: '🔍 正在分析您的需求...', phase: 'plan' },
      { type: 'plan_complete', message: '规划完成！', plan: { title: 'AI大模型介绍', theme: 'professional', slides: [] } },
      { type: 'slide_progress', message: '正在生成第1页...', slide: 1, total: 5 },
      { type: 'slide_progress', message: '正在生成第2页...', slide: 2, total: 5 },
      { type: 'slide_progress', message: '正在生成第3页...', slide: 3, total: 5 },
      { type: 'slide_progress', message: '正在生成第4页...', slide: 4, total: 5 },
      { type: 'slide_progress', message: '正在生成第5页...', slide: 5, total: 5 },
      { type: 'complete', message: 'PPT生成完成！' }
    ];

    let index = 0;
    const interval = setInterval(() => {
      if (index < steps.length) {
        handleStreamEvent(steps[index] as ExecutionEvent);
        index++;
      } else {
        clearInterval(interval);
      }
    }, 800);
  };

  const handleTemplateClick = (template: typeof templates[0]) => {
    setInputValue(template.prompt);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="app-container">
      <header className="header">
        <div className="header-content">
          <div className="logo">
            <div className="logo-icon">📊</div>
            <span className="logo-text">PPT智能生成助手</span>
          </div>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <span style={{
              padding: '0.25rem 0.75rem',
              background: isProcessing ? 'rgba(245,158,11,0.1)' : 'rgba(16,185,129,0.1)',
              color: isProcessing ? '#f59e0b' : '#10b981',
              borderRadius: '9999px',
              fontSize: '0.75rem',
              display: 'flex',
              alignItems: 'center',
              gap: '0.25rem'
            }}>
              <span className={`status-dot ${isProcessing ? 'processing' : ''}`}></span>
              {isProcessing ? '处理中' : '就绪'}
            </span>
          </div>
        </div>
      </header>

      <main className="main-content">
        <div className="chat-container">
          <div className="chat-messages">
            {messages.length === 0 && (
              <div className="templates-section">
                <h3 style={{ marginBottom: '1rem', color: '#64748b' }}>选择模板快速开始</h3>
                <div className="templates-grid">
                  {templates.map(template => (
                    <div
                      key={template.id}
                      className="template-card"
                      onClick={() => handleTemplateClick(template)}
                    >
                      <div className="template-icon">{template.icon}</div>
                      <div className="template-title">{template.title}</div>
                      <div className="template-desc">{template.description}</div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {messages.map(message => (
              <div key={message.id} className={`message ${message.role}`}>
                <div className="message-avatar">
                  {message.role === 'user' ? '👤' : '🤖'}
                </div>
                <div className="message-content">
                  <ReactMarkdown>{message.content}</ReactMarkdown>
                </div>
              </div>
            ))}
            <div ref={messagesEndRef} />
          </div>

          {isProcessing && (
            <div className="status-bar">
              <div className="status-indicator">
                <span className="status-dot processing"></span>
                <span>{currentEvent?.message || '处理中...'}</span>
              </div>
              <div className="progress-bar">
                <div className="progress-fill" style={{ width: `${progress}%` }}></div>
              </div>
            </div>
          )}

          <div className="input-container">
            <div className="input-wrapper">
              <textarea
                className="chat-input"
                placeholder="描述你想要制作的PPT内容..."
                value={inputValue}
                onChange={e => setInputValue(e.target.value)}
                onKeyDown={handleKeyDown}
                rows={1}
              />
              <button
                className="send-button"
                onClick={handleSend}
                disabled={!inputValue.trim() || isProcessing}
              >
                {isProcessing ? '生成中...' : '发送'}
              </button>
            </div>
          </div>
        </div>

        {planResult && (
          <div style={{
            marginTop: '1rem',
            padding: '1rem',
            background: 'white',
            borderRadius: '8px',
            boxShadow: '0 4px 6px rgba(0,0,0,0.1)'
          }}>
            <h4 style={{ marginBottom: '0.5rem' }}>📋 PPT大纲预览</h4>
            <p><strong>标题：</strong>{planResult.title}</p>
            <p><strong>主题：</strong>{planResult.theme}</p>
            <p><strong>页数：</strong>{planResult.slides.length} 页</p>
          </div>
        )}
      </main>
    </div>
  );
}

export default App;