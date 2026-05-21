<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { authState } from '../stores/auth';
import { isLoggedIn } from '../api';

const router = useRouter();
const auth = authState;

const featuresRef = ref<HTMLElement | null>(null);

function scrollToFeatures() {
  featuresRef.value?.scrollIntoView({ behavior: 'smooth' });
}

function handleGetStarted() {
  if (isLoggedIn()) {
    router.push('/dashboard');
  } else {
    router.push('/auth');
  }
}

onMounted(async () => {
  if (isLoggedIn()) await auth.init();
});
</script>

<template>
  <div class="home">
    <!-- ═══════════════════════════════════════════════════════════ -->
    <!-- HERO (full viewport)                                        -->
    <!-- ═══════════════════════════════════════════════════════════ -->
    <section class="hero">
      <div class="hero-bg">
        <div class="bg-grid"></div>
        <div class="bg-orb orb-1"></div>
        <div class="bg-orb orb-2"></div>
        <div class="bg-orb orb-3"></div>
      </div>

      <div class="hero-content">
        <div class="hero-badge">
          <span class="badge-dot"></span>
          AI-Powered Presentation Generator
        </div>

        <h1 class="hero-title">
          <span class="title-line">用 AI 重新定义</span>
          <span class="title-line accent">PPT 制作</span>
        </h1>

        <p class="hero-sub">
          只需一句话描述你的需求，智能体自动规划大纲、生成幻灯片、视觉质检、迭代修复<br/>
          从想法到专业演示文稿，最快仅需 <strong>2 分钟</strong>
        </p>

        <div class="hero-actions">
          <button class="hero-btn primary" @click="handleGetStarted">
            {{ auth.loggedIn ? '进入工作台' : '立即开始' }}
            <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2"><path d="M7 4l8 6-8 6"/></svg>
          </button>
          <button class="hero-btn secondary" @click="scrollToFeatures">
            了解更多
            <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2"><path d="M7 8l3 3 3-3"/></svg>
          </button>
        </div>

        <div class="hero-stats">
          <div class="stat">
            <span class="stat-num">3-5x</span>
            <span class="stat-label">并行加速比</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat">
            <span class="stat-num">AI</span>
            <span class="stat-label">全自动质检</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat">
            <span class="stat-num">36+</span>
            <span class="stat-label">内置主题模板</span>
          </div>
        </div>
      </div>

      <div class="scroll-indicator" @click="scrollToFeatures">
        <span class="scroll-text">向下滚动</span>
        <div class="scroll-mouse">
          <div class="scroll-wheel"></div>
        </div>
      </div>
    </section>

    <!-- ═══════════════════════════════════════════════════════════ -->
    <!-- FEATURES                                                   -->
    <!-- ═══════════════════════════════════════════════════════════ -->
    <section ref="featuresRef" class="features">
      <div class="features-header">
        <h2 class="section-title">为什么选择 PPT Agent</h2>
        <p class="section-sub">不是简单的模板填充 — 而是从内容到设计的全链路 AI 生成</p>
      </div>

      <div class="features-grid">
        <div class="feature-card">
          <div class="feature-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/>
            </svg>
          </div>
          <h3>智能并行生成</h3>
          <p>多张幻灯片同时生成，效率提升 3-5 倍。智能依赖分析保证内容连贯性。</p>
          <ul class="feature-details">
            <li>自适应并发控制</li>
            <li>内容逻辑依赖管理</li>
            <li>实时追踪每页状态</li>
          </ul>
        </div>

        <div class="feature-card">
          <div class="feature-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
          </div>
          <h3>视觉质量审查</h3>
          <p>自动将 PPTX 渲染为图片，多模态 LLM 审查排版、对齐、配色。</p>
          <ul class="feature-details">
            <li>PPTX→PDF→JPEG 管线</li>
            <li>多模态视觉审查</li>
            <li>自动修复 ≤ 2 次迭代</li>
          </ul>
        </div>

        <div class="feature-card">
          <div class="feature-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <rect x="3" y="3" width="18" height="18" rx="2"/>
              <line x1="7" y1="8" x2="17" y2="8"/>
              <line x1="7" y1="12" x2="17" y2="12"/>
              <line x1="7" y1="16" x2="12" y2="16"/>
            </svg>
          </div>
          <h3>结构化内容设计</h3>
          <p>从大纲规划到元素排版，全自动处理标题、正文、图表、图片布局。</p>
          <ul class="feature-details">
            <li>智能大纲生成</li>
            <li>36 套内置主题</li>
            <li>多内容类型支持</li>
          </ul>
        </div>
      </div>

      <!-- CTA -->
      <div class="cta-banner">
        <div class="cta-glow"></div>
        <h2>准备好提升你的 PPT 效率了吗？</h2>
        <p>无需学习复杂操作，只需用自然语言描述想法</p>
        <button class="hero-btn primary large" @click="handleGetStarted">
          {{ auth.loggedIn ? '进入工作台' : '免费开始使用' }}
          <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2"><path d="M7 4l8 6-8 6"/></svg>
        </button>
      </div>
    </section>

    <!-- ═══════════════════════════════════════════════════════════ -->
    <!-- FOOTER                                                     -->
    <!-- ═══════════════════════════════════════════════════════════ -->
    <footer class="footer">
      <div class="footer-content">
        <div class="footer-brand">
          <svg viewBox="0 0 40 40" fill="none" class="footer-logo">
            <rect x="4" y="6" width="32" height="24" rx="3" fill="var(--accent-soft)"/>
            <rect x="4" y="6" width="32" height="6" rx="3" fill="var(--accent)"/>
            <rect x="10" y="16" width="14" height="2" rx="1" fill="var(--accent)"/>
            <rect x="10" y="20" width="20" height="2" rx="1" fill="var(--border)"/>
            <rect x="10" y="24" width="16" height="2" rx="1" fill="var(--border)"/>
          </svg>
          <span>PPT Agent</span>
        </div>
        <p class="footer-copy">Powered by Eino ADK &amp; Multi-Agent Architecture</p>
      </div>
    </footer>
  </div>
</template>

<style scoped>
/* ═══════════════════════════════════════════════════════════════════════ */
/* HERO — Clean Light Modern                                         */
/* ═══════════════════════════════════════════════════════════════════════ */
.hero {
  min-height: 100vh;
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  position: relative; overflow: hidden;
  background: var(--bg-base);
}

/* Top gradient accent bar */
.hero::before {
  content: '';
  position: absolute; top: 0; left: 0; right: 0; height: 4px;
  background: linear-gradient(90deg, var(--accent), #8b5cf6, var(--info));
}

.hero-bg { position: absolute; inset: 0; pointer-events: none; }

.bg-grid {
  position: absolute; inset: 0;
  background-image:
    linear-gradient(rgba(99,102,241,0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(99,102,241,0.03) 1px, transparent 1px);
  background-size: 60px 60px;
}

.bg-orb { position: absolute; border-radius: 50%; filter: blur(80px); opacity: 0.07; }
.orb-1 { width: 500px; height: 500px; background: var(--accent); top: -200px; right: -100px; animation: float 12s ease-in-out infinite; }
.orb-2 { width: 350px; height: 350px; background: var(--info); bottom: -100px; left: -50px; animation: float 10s ease-in-out infinite reverse; }
.orb-3 { width: 280px; height: 280px; background: #8b5cf6; top: 40%; left: 50%; transform: translate(-50%,-50%); animation: float 14s ease-in-out infinite 3s; }

.hero-content { position: relative; z-index: 1; text-align: center; max-width: 760px; padding: 0 2rem; }

.hero-badge {
  display: inline-flex; align-items: center; gap: 0.5rem;
  padding: 0.35rem 0.9rem;
  border: 1px solid var(--accent-border);
  border-radius: var(--radius-full);
  background: var(--accent-soft);
  color: var(--accent); font-size: 0.75rem; font-weight: 500;
  margin-bottom: 2rem;
  letter-spacing: 0.01em;
}
.badge-dot { width: 6px; height: 6px; border-radius: 50%; background: var(--accent); animation: pulse 2s infinite; }

.hero-title { font-size: 3.5rem; font-weight: 800; line-height: 1.15; margin-bottom: 1.5rem; letter-spacing: -0.025em; }
.title-line { display: block; color: var(--text); }
.title-line.accent {
  background: linear-gradient(135deg, var(--accent) 0%, #8b5cf6 100%);
  -webkit-background-clip: text; -webkit-text-fill-color: transparent;
  font-size: 4rem;
}

.hero-sub { font-size: 1.05rem; color: var(--text-secondary); line-height: 1.7; margin-bottom: 2.5rem; max-width: 560px; margin-left: auto; margin-right: auto; }
.hero-sub strong { color: var(--accent); font-weight: 600; }

.hero-actions { display: flex; gap: 0.75rem; justify-content: center; margin-bottom: 3rem; }

.hero-btn {
  display: inline-flex; align-items: center; gap: 0.4rem;
  padding: 0.8rem 1.6rem; border-radius: var(--radius-md);
  font-size: 0.9rem; font-weight: 600; cursor: pointer; border: none;
  transition: all var(--transition-md);
  font-family: inherit;
  text-decoration: none;
}
.hero-btn svg { width: 16px; height: 16px; transition: transform var(--transition-md); }
.hero-btn:hover svg { transform: translateX(3px); }
.hero-btn.primary {
  background: var(--accent); color: #fff;
  box-shadow: 0 4px 14px rgba(99, 102, 241, 0.3);
}
.hero-btn.primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(99, 102, 241, 0.4);
  background: var(--accent-hover);
}
.hero-btn.secondary {
  background: var(--bg-base); color: var(--text-secondary);
  border: 1.5px solid var(--border);
  box-shadow: var(--shadow-xs);
}
.hero-btn.secondary:hover { background: var(--bg-muted); border-color: var(--accent-border); color: var(--text); }
.hero-btn.large { padding: 0.95rem 2rem; font-size: 0.95rem; }

.hero-stats { display: flex; align-items: center; gap: 2.5rem; justify-content: center; }
.stat { text-align: center; }
.stat-num { display: block; font-size: 1.4rem; font-weight: 800; color: var(--text); letter-spacing: -0.02em; }
.stat-label { font-size: 0.7rem; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.06em; font-weight: 500; }
.stat-divider { width: 1px; height: 28px; background: var(--border); }

.scroll-indicator {
  position: absolute; bottom: 2rem; left: 50%; transform: translateX(-50%);
  display: flex; flex-direction: column; align-items: center; gap: 0.4rem;
  cursor: pointer; z-index: 1;
}
.scroll-text { font-size: 0.62rem; text-transform: uppercase; letter-spacing: 0.1em; color: var(--text-muted); }
.scroll-mouse { width: 22px; height: 36px; border: 1.5px solid var(--border); border-radius: 11px; display: flex; justify-content: center; padding-top: 6px; }
.scroll-wheel { width: 3px; height: 7px; background: var(--accent); border-radius: 2px; animation: scrollWheel 2s ease-in-out infinite; }

/* ═══════════════════════════════════════════════════════════════════════ */
/* FEATURES — Clean Light Cards                                       */
/* ═══════════════════════════════════════════════════════════════════════ */
.features { background: var(--bg-muted); padding: 5rem 2rem; }

.features-header { text-align: center; margin-bottom: 3.5rem; }
.section-title { font-size: 2rem; font-weight: 700; color: var(--text); margin-bottom: 0.75rem; letter-spacing: -0.02em; }
.section-sub { font-size: 0.95rem; color: var(--text-secondary); }

.features-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1.5rem; max-width: 960px; margin: 0 auto 5rem; }

.feature-card {
  background: var(--bg-base); border: 1px solid var(--border);
  border-radius: var(--radius-lg); padding: 1.75rem 1.5rem;
  transition: transform var(--transition-md), box-shadow var(--transition-md), border-color var(--transition-md);
  position: relative; overflow: hidden;
  box-shadow: var(--shadow-card);
}
.feature-card::before {
  content: ''; position: absolute; top: 0; left: 0; right: 0; height: 2px;
  background: linear-gradient(90deg, transparent, var(--accent), transparent);
  opacity: 0; transition: opacity var(--transition-md);
}
.feature-card:hover {
  transform: translateY(-4px); border-color: var(--accent-border);
  box-shadow: var(--shadow-card-hover);
}
.feature-card:hover::before { opacity: 1; }

.feature-icon {
  width: 44px; height: 44px; border-radius: var(--radius-md);
  background: var(--accent-soft); color: var(--accent);
  display: flex; align-items: center; justify-content: center; margin-bottom: 1.1rem;
  box-shadow: 0 0 0 4px rgba(99,102,241,0.06);
}
.feature-icon svg { width: 22px; height: 22px; }
.feature-card h3 { font-size: 1rem; font-weight: 700; color: var(--text); margin-bottom: 0.5rem; }
.feature-card > p { font-size: 0.82rem; color: var(--text-secondary); line-height: 1.65; margin-bottom: 1rem; }

.feature-details { list-style: none; padding: 0; }
.feature-details li { font-size: 0.75rem; color: var(--text-secondary); padding: 0.2rem 0 0.2rem 1rem; position: relative; }
.feature-details li::before { content: ''; position: absolute; left: 0; top: 50%; transform: translateY(-50%); width: 4px; height: 4px; border-radius: 50%; background: var(--accent); }

/* CTA BANNER */
.cta-banner {
  max-width: 760px; margin: 0 auto; text-align: center;
  background: var(--bg-base); border: 1px solid var(--border);
  border-radius: var(--radius-xl); padding: 3rem 2rem;
  position: relative; overflow: hidden;
  box-shadow: var(--shadow-card);
  display: flex; flex-direction: column; align-items: center;
}
.cta-banner::before {
  content: ''; position: absolute; top: 0; left: 0; right: 0; height: 3px;
  background: linear-gradient(90deg, var(--accent), #8b5cf6);
}
.cta-glow { position: absolute; width: 300px; height: 200px; background: var(--accent); border-radius: 50%; filter: blur(80px); opacity: 0.05; top: -50px; left: 50%; transform: translateX(-50%); }
.cta-banner h2 { font-size: 1.5rem; font-weight: 700; color: var(--text); margin-bottom: 0.5rem; letter-spacing: -0.01em; position: relative; }
.cta-banner p { color: var(--text-secondary); margin-bottom: 1.5rem; font-size: 0.9rem; position: relative; }

/* FOOTER */
.footer { background: var(--bg-base); border-top: 1px solid var(--border); padding: 2rem; text-align: center; }
.footer-brand { display: flex; align-items: center; justify-content: center; gap: 0.5rem; margin-bottom: 0.5rem; }
.footer-logo { width: 26px; height: 26px; }
.footer-brand span { font-size: 0.85rem; font-weight: 600; color: var(--text); }
.footer-copy { font-size: 0.7rem; color: var(--text-muted); }

/* KEYFRAMES */
@keyframes float { 0%, 100% { transform: translateY(0) scale(1); } 50% { transform: translateY(-35px) scale(1.03); } }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.4; } }
@keyframes scrollWheel { 0%, 100% { transform: translateY(0); opacity: 1; } 50% { transform: translateY(5px); opacity: 0.3; } }

/* RESPONSIVE */
@media (max-width: 768px) {
  .hero-title { font-size: 2.4rem; }
  .title-line.accent { font-size: 2.6rem; }
  .hero-sub { font-size: 0.9rem; }
  .hero-actions { flex-direction: column; align-items: center; }
  .hero-stats { gap: 1.5rem; flex-wrap: wrap; }
  .stat-divider { display: none; }
  .features-grid { grid-template-columns: 1fr; }
  .features { padding: 3rem 1.25rem; }
  .feature-card { padding: 1.25rem 1rem; }
  .section-title { font-size: 1.6rem; }
  .cta-banner { padding: 2rem 1.25rem; }
}
</style>
