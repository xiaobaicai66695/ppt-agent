<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
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
        <div class="bg-glow glow-1"></div>
        <div class="bg-glow glow-2"></div>
        <div class="bg-glow glow-3"></div>
        <div class="bg-particles">
          <span v-for="n in 20" :key="n" class="particle" :style="{
            left: ((n * 137 + 53) % 100) + '%',
            top: ((n * 73 + 27) % 100) + '%',
            animationDelay: (n * 0.7) + 's',
            animationDuration: (3 + n % 4) + 's',
            width: (3 + n % 4) + 'px',
            height: (3 + n % 4) + 'px',
          }"></span>
        </div>
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
            <rect x="4" y="6" width="32" height="24" rx="3" fill="rgba(255,255,255,0.15)"/>
            <rect x="4" y="6" width="32" height="5" rx="3" fill="rgba(255,255,255,0.3)"/>
            <rect x="10" y="14" width="14" height="2" rx="1" fill="rgba(255,255,255,0.2)"/>
            <rect x="10" y="18" width="20" height="2" rx="1" fill="rgba(255,255,255,0.15)"/>
          </svg>
          <span>PPT Agent</span>
        </div>
        <p class="footer-copy">Powered by Eino ADK &amp; Multi-Agent Architecture</p>
      </div>
    </footer>
  </div>
</template>

<style scoped>
/* ═══════════════════════════════════════════════════════════════ */
/* HERO                                                           */
/* ═══════════════════════════════════════════════════════════════ */
.hero {
  min-height: 100vh;
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  position: relative; overflow: hidden;
  background: #0b1121;
}

.hero-bg { position: absolute; inset: 0; pointer-events: none; }

.bg-grid {
  position: absolute; inset: 0;
  background-image:
    linear-gradient(rgba(99,102,241,0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(99,102,241,0.03) 1px, transparent 1px);
  background-size: 60px 60px;
}

.bg-glow { position: absolute; border-radius: 50%; filter: blur(120px); opacity: 0.12; }
.glow-1 { width: 600px; height: 600px; background: #6366f1; top: -200px; right: -200px; animation: floatSlow 12s ease-in-out infinite; }
.glow-2 { width: 400px; height: 400px; background: #3b82f6; bottom: -100px; left: -100px; animation: floatSlow 10s ease-in-out infinite reverse; }
.glow-3 { width: 350px; height: 350px; background: #8b5cf6; top: 50%; left: 50%; transform: translate(-50%, -50%); animation: floatSlow 14s ease-in-out infinite 3s; }

.bg-particles { position: absolute; inset: 0; }
.particle {
  position: absolute;
  background: rgba(129,140,248,0.4);
  border-radius: 50%;
  animation: particleFloat linear infinite;
}

.hero-content { position: relative; z-index: 1; text-align: center; max-width: 800px; padding: 0 2rem; }

.hero-badge {
  display: inline-flex; align-items: center; gap: 0.5rem;
  padding: 0.4rem 1rem;
  border: 1px solid rgba(129,140,248,0.3);
  border-radius: 999px;
  background: rgba(99,102,241,0.08);
  color: #a5b4fc; font-size: 0.78rem; font-weight: 500;
  margin-bottom: 2rem;
  backdrop-filter: blur(12px);
}
.badge-dot { width: 7px; height: 7px; border-radius: 50%; background: #818cf8; animation: pulse 2s infinite; }

.hero-title { font-size: 3.5rem; font-weight: 800; line-height: 1.15; margin-bottom: 1.5rem; letter-spacing: -0.02em; }
.title-line { display: block; background: linear-gradient(135deg, #e2e8f0 0%, #cbd5e1 100%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
.title-line.accent { background: linear-gradient(135deg, #818cf8 0%, #6366f1 40%, #8b5cf6 100%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; font-size: 4rem; }

.hero-sub { font-size: 1.05rem; color: #64748b; line-height: 1.7; margin-bottom: 2.5rem; max-width: 580px; margin-left: auto; margin-right: auto; }
.hero-sub strong { color: #818cf8; font-weight: 600; }

.hero-actions { display: flex; gap: 1rem; justify-content: center; margin-bottom: 3rem; }

.hero-btn {
  display: inline-flex; align-items: center; gap: 0.4rem;
  padding: 0.85rem 1.75rem; border-radius: 12px;
  font-size: 0.95rem; font-weight: 600; cursor: pointer; border: none;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  font-family: inherit;
}
.hero-btn svg { width: 18px; height: 18px; transition: transform 0.25s; }
.hero-btn:hover svg { transform: translateX(3px); }
.hero-btn.primary { background: linear-gradient(135deg, #6366f1, #8b5cf6); color: #fff; box-shadow: 0 4px 20px rgba(99,102,241,0.3); }
.hero-btn.primary:hover { transform: translateY(-2px); box-shadow: 0 8px 30px rgba(99,102,241,0.45); }
.hero-btn.secondary { background: rgba(255,255,255,0.04); color: #94a3b8; border: 1px solid rgba(255,255,255,0.1); }
.hero-btn.secondary:hover { background: rgba(255,255,255,0.08); color: #cbd5e1; }
.hero-btn.large { padding: 1rem 2.25rem; font-size: 1rem; }

.hero-stats { display: flex; align-items: center; gap: 2rem; justify-content: center; }
.stat { text-align: center; }
.stat-num { display: block; font-size: 1.3rem; font-weight: 700; color: #e2e8f0; }
.stat-label { font-size: 0.72rem; color: #475569; text-transform: uppercase; letter-spacing: 0.06em; }
.stat-divider { width: 1px; height: 28px; background: rgba(255,255,255,0.06); }

.scroll-indicator {
  position: absolute; bottom: 2rem; left: 50%; transform: translateX(-50%);
  display: flex; flex-direction: column; align-items: center; gap: 0.5rem;
  cursor: pointer; z-index: 1;
}
.scroll-text { font-size: 0.65rem; text-transform: uppercase; letter-spacing: 0.1em; color: #475569; }
.scroll-mouse { width: 22px; height: 36px; border: 2px solid rgba(255,255,255,0.1); border-radius: 11px; display: flex; justify-content: center; padding-top: 6px; }
.scroll-wheel { width: 3px; height: 8px; background: #6366f1; border-radius: 2px; animation: scrollWheel 2s ease-in-out infinite; }

/* ═══════════════════════════════════════════════════════════════ */
/* FEATURES                                                       */
/* ═══════════════════════════════════════════════════════════════ */
.features { background: #0f172a; padding: 5rem 2rem; }

.features-header { text-align: center; margin-bottom: 3.5rem; }
.section-title { font-size: 2.2rem; font-weight: 700; color: #f1f5f9; margin-bottom: 0.75rem; }
.section-sub { font-size: 1rem; color: #64748b; }

.features-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1.5rem; max-width: 1000px; margin: 0 auto 5rem; }

.feature-card {
  background: rgba(255,255,255,0.02); border: 1px solid rgba(255,255,255,0.06);
  border-radius: 16px; padding: 2rem 1.5rem;
  transition: transform 0.3s, border-color 0.3s, box-shadow 0.3s;
  position: relative; overflow: hidden;
}
.feature-card::before { content: ''; position: absolute; top: 0; left: 0; right: 0; height: 1px; background: linear-gradient(90deg, transparent, rgba(99,102,241,0.3), transparent); opacity: 0; transition: opacity 0.3s; }
.feature-card:hover { transform: translateY(-4px); border-color: rgba(99,102,241,0.2); box-shadow: 0 12px 40px rgba(0,0,0,0.2); }
.feature-card:hover::before { opacity: 1; }

.feature-icon { width: 48px; height: 48px; border-radius: 12px; background: rgba(99,102,241,0.1); color: #818cf8; display: flex; align-items: center; justify-content: center; margin-bottom: 1.25rem; }
.feature-icon svg { width: 24px; height: 24px; }
.feature-card h3 { font-size: 1.1rem; font-weight: 600; color: #e2e8f0; margin-bottom: 0.6rem; }
.feature-card > p { font-size: 0.82rem; color: #64748b; line-height: 1.6; margin-bottom: 1rem; }

.feature-details { list-style: none; padding: 0; }
.feature-details li { font-size: 0.75rem; color: #475569; padding: 0.25rem 0 0.25rem 1rem; position: relative; }
.feature-details li::before { content: ''; position: absolute; left: 0; top: 50%; transform: translateY(-50%); width: 4px; height: 4px; border-radius: 50%; background: #6366f1; }

.cta-banner {
  max-width: 800px; margin: 0 auto; text-align: center;
  background: rgba(99,102,241,0.04); border: 1px solid rgba(99,102,241,0.1);
  border-radius: 20px; padding: 3.5rem 2rem;
  position: relative; overflow: hidden;
}
.cta-glow { position: absolute; width: 300px; height: 300px; background: #6366f1; border-radius: 50%; filter: blur(120px); opacity: 0.08; top: -100px; left: 50%; transform: translateX(-50%); }
.cta-banner h2 { font-size: 1.6rem; font-weight: 700; color: #f1f5f9; margin-bottom: 0.5rem; position: relative; }
.cta-banner p { color: #64748b; margin-bottom: 1.5rem; position: relative; }

/* ═══════════════════════════════════════════════════════════════ */
/* FOOTER                                                         */
/* ═══════════════════════════════════════════════════════════════ */
.footer { background: #0b1121; border-top: 1px solid rgba(255,255,255,0.04); padding: 2.5rem 2rem; text-align: center; }
.footer-brand { display: flex; align-items: center; justify-content: center; gap: 0.5rem; margin-bottom: 0.5rem; }
.footer-logo { width: 28px; height: 28px; }
.footer-brand span { font-size: 0.9rem; font-weight: 600; color: #cbd5e1; }
.footer-copy { font-size: 0.72rem; color: #475569; }

/* ═══════════════════════════════════════════════════════════════ */
/* KEYFRAMES                                                      */
/* ═══════════════════════════════════════════════════════════════ */
@keyframes floatSlow {
  0%, 100% { transform: translateY(0) scale(1); }
  50% { transform: translateY(-40px) scale(1.05); }
}
@keyframes particleFloat {
  0% { transform: translateY(0) translateX(0); opacity: 0; }
  10% { opacity: 1; }
  90% { opacity: 1; }
  100% { transform: translateY(-100vh) translateX(40px); opacity: 0; }
}
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}
@keyframes scrollWheel {
  0%, 100% { transform: translateY(0); opacity: 1; }
  50% { transform: translateY(6px); opacity: 0.3; }
}

/* ═══════════════════════════════════════════════════════════════ */
/* RESPONSIVE                                                     */
/* ═══════════════════════════════════════════════════════════════ */
@media (max-width: 768px) {
  .hero-title { font-size: 2.2rem; }
  .title-line.accent { font-size: 2.5rem; }
  .hero-sub { font-size: 0.9rem; }
  .hero-actions { flex-direction: column; align-items: center; }
  .hero-stats { gap: 1rem; flex-wrap: wrap; }
  .stat-divider { display: none; }
  .features-grid { grid-template-columns: 1fr; }
  .section-title { font-size: 1.6rem; }
}
</style>
