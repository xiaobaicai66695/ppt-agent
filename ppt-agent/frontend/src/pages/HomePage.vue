<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowRight, ArrowUpRight, Check, FileStack, Sparkles } from 'lucide-vue-next'
import PPTStack from '../components/PPTStack.vue'
import { isLoggedIn } from '../api'

const router = useRouter()
const brief = ref('')
const focused = ref(false)
const hoveringHero = ref(false)
const hero = ref<HTMLElement | null>(null)
const workflow = ref<HTMLElement | null>(null)
const canStart = computed(() => brief.value.trim().length > 0)

let animationFrame: number | undefined
const targetX = ref(52)
const targetY = ref(46)
let observer: IntersectionObserver | undefined

function start() {
  const request = brief.value.trim()
  router.push(isLoggedIn()
    ? { path: '/dashboard', query: { brief: request } }
    : { path: '/auth', query: { next: `/dashboard?brief=${encodeURIComponent(request)}` } })
}

function writeMotion() {
  animationFrame = undefined
  if (!hero.value) return
  hero.value.style.setProperty('--pointer-x', `${targetX.value}%`)
  hero.value.style.setProperty('--pointer-y', `${targetY.value}%`)
  hero.value.style.setProperty('--tilt-x', `${(targetX.value - 50) / 4}px`)
  hero.value.style.setProperty('--tilt-y', `${(targetY.value - 50) / 5}px`)
}

function moveHero(event: PointerEvent) {
  if (!hero.value || window.matchMedia('(prefers-reduced-motion: reduce)').matches) return
  const bounds = hero.value.getBoundingClientRect()
  targetX.value = Math.max(0, Math.min(100, ((event.clientX - bounds.left) / bounds.width) * 100))
  targetY.value = Math.max(0, Math.min(100, ((event.clientY - bounds.top) / bounds.height) * 100))
  if (!animationFrame) animationFrame = window.requestAnimationFrame(writeMotion)
}

function resetHero() {
  hoveringHero.value = false
  targetX.value = 52
  targetY.value = 46
  if (!animationFrame) animationFrame = window.requestAnimationFrame(writeMotion)
}

function scrollToWorkflow() {
  workflow.value?.scrollIntoView({ behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth', block: 'start' })
}

onMounted(() => {
  observer = new IntersectionObserver(entries => {
    entries.forEach(entry => entry.target.classList.toggle('is-visible', entry.isIntersecting))
  }, { threshold: 0.18 })
  document.querySelectorAll<HTMLElement>('[data-reveal]').forEach(element => observer?.observe(element))
})

onBeforeUnmount(() => {
  if (animationFrame) window.cancelAnimationFrame(animationFrame)
  observer?.disconnect()
})
</script>

<template>
  <main class="landing">
    <header class="home-nav">
      <RouterLink to="/" class="wordmark"><span><FileStack :size="17" /></span>PPTform</RouterLink>
      <nav aria-label="首页导航">
        <button type="button" class="text-link" @click="scrollToWorkflow">如何工作</button>
        <RouterLink class="text-link" to="/auth">登录</RouterLink>
        <RouterLink class="small-button" to="/auth">开始创作 <ArrowUpRight :size="15" /></RouterLink>
      </nav>
    </header>

    <section ref="hero" class="hero" :class="{ 'is-hovering': hoveringHero }" @pointermove="moveHero" @pointerenter="hoveringHero = true" @pointerleave="resetHero">
      <div class="hero-atmosphere" aria-hidden="true">
        <i class="mist mist-one"></i><i class="mist mist-two"></i><i class="mist mist-three"></i>
        <i class="pointer-light"></i><i class="grain"></i>
      </div>
      <div class="hero-copy" data-reveal>
        <p class="eyeline"><i></i>把思考变成一场好演示</p>
        <h1>让每一个想法<br>都有展开的方式。</h1>
        <p class="lede">输入主题或一段资料。PPTform 会完成研究、结构规划和页面编排，交付一份真正能讲清楚事情的演示文稿。</p>
        <form class="brief-box" novalidate :class="{ focused }" @submit.prevent="start">
          <label class="sr-only" for="home-brief">演示需求</label>
          <textarea id="home-brief" class="resize-none" v-model="brief" rows="2" placeholder="例如：为新能源储能方案做一份 10 页客户提案" @focus="focused = true" @blur="focused = false" />
          <button :disabled="!canStart" type="submit" aria-label="开始生成演示"><ArrowRight :size="21" /></button>
          <span>可直接输入需求，也可稍后编排每一页</span>
        </form>
        <div class="proof"><span><Check :size="14" />先规划，再生成</span><span><Check :size="14" />过程随时可见</span><span><Check :size="14" />导出可编辑 PPTX</span></div>
      </div>
      <div class="hero-object" aria-hidden="true">
        <p class="object-caption">从问题到叙事 / 01</p>
        <PPTStack :active="focused || hoveringHero" :tilt-x="(targetX - 52) / 4" :tilt-y="(targetY - 46) / 5" />
      </div>
      <button type="button" class="scroll-cue" @click="scrollToWorkflow"><span>继续探索</span><i></i></button>
    </section>

    <section ref="workflow" class="workflow" aria-labelledby="workflow-title">
      <div class="workflow-heading" data-reveal>
        <p class="eyeline"><i></i>不是模板堆砌，是一条清晰的创作路径</p>
        <h2 id="workflow-title">从一个问题，<br>到一份可讲述的作品。</h2>
      </div>
      <div class="sequence">
        <article data-reveal><b>01</b><div><h3>给出问题</h3><p>一句需求、一段资料，或一份已有大纲。创作从你已经拥有的内容开始。</p></div><span class="sequence-mark">提问</span></article>
        <article data-reveal><b>02</b><div><h3>看它成形</h3><p>研究、结构、视觉与每一页的交付都有可查看的轨迹，不让进度藏在黑箱里。</p></div><span class="sequence-mark">铺陈</span></article>
        <article data-reveal><b>03</b><div><h3>继续编辑</h3><p>预览、下载、继续修订。成品属于你的工作流，而不是一个一次性的答案。</p></div><span class="sequence-mark">讲述</span></article>
      </div>
    </section>

    <section class="final" data-reveal>
      <div><p class="eyeline"><i></i>一套为演示而生的工作流</p><h2>你的下一份演示，<br>从一个念头开始。</h2></div>
      <RouterLink class="final-cta" to="/auth"><Sparkles :size="18" />开始一份演示</RouterLink>
    </section>
  </main>
</template>

<style scoped>
.landing{min-height:100vh;overflow:hidden;color:#173b42;background:#f7f4ed}.home-nav{position:relative;z-index:4;height:78px;display:flex;align-items:center;justify-content:space-between;width:min(1220px,calc(100% - 48px));margin:auto;border-bottom:1px solid rgba(32,65,62,.13)}.home-nav nav{display:flex;align-items:center;gap:22px}.wordmark{display:flex;align-items:center;gap:9px;font:700 20px 'Noto Serif SC',serif;letter-spacing:-.045em}.wordmark span{display:grid;place-items:center;width:27px;height:27px;background:#1c6164;color:#f4f4e9;border-radius:8px 8px 2px 8px}.text-link{padding:5px 0;border:0;color:#52716e;background:transparent;font-size:13px}.text-link:hover{color:#143e44}.small-button,.final-cta{display:flex;align-items:center;gap:8px;padding:10px 13px;color:#fffdf7;background:#d56d2d;border-radius:6px;font-size:13px;font-weight:700;transition:transform .2s,background .2s}.small-button:hover,.final-cta:hover{transform:translateY(-2px);background:#b94f21}
.hero{--pointer-x:52%;--pointer-y:46%;--tilt-x:0px;--tilt-y:0px;position:relative;display:grid;grid-template-columns:minmax(0,1.03fr) minmax(310px,.97fr);align-items:center;gap:30px;width:min(1220px,calc(100% - 48px));min-height:635px;margin:auto;padding:85px 0 100px;isolation:isolate}.hero-atmosphere{position:absolute;z-index:-1;inset:-78px min(-21vw,-220px) -35px;overflow:hidden;pointer-events:none;background:#f3ded1}.mist{position:absolute;display:block;width:50%;aspect-ratio:1;border-radius:48% 52% 58% 42%;filter:blur(30px);opacity:.7;transition:transform .55s cubic-bezier(.2,.75,.25,1)}.mist-one{top:-22%;left:-10%;background:radial-gradient(circle at 58% 54%,#e2a47f 0 7%,#edc2aa 29%,transparent 66%);transform:translate(calc(var(--tilt-x) * -.6),calc(var(--tilt-y) * -.45)) rotate(-18deg)}.mist-two{right:-11%;top:3%;background:radial-gradient(ellipse at 46% 43%,#e9b78f 0 9%,#f4d7c3 38%,transparent 70%);transform:translate(calc(var(--tilt-x) * .8),calc(var(--tilt-y) * .55)) rotate(22deg)}.mist-three{width:63%;left:28%;bottom:-46%;background:radial-gradient(ellipse at 50% 30%,#e2a179 0 4%,#efc6ad 32%,transparent 68%);transform:translate(calc(var(--tilt-x) * .4),calc(var(--tilt-y) * .75)) rotate(-15deg)}.pointer-light{position:absolute;width:420px;aspect-ratio:1;border-radius:50%;left:calc(var(--pointer-x) - 210px);top:calc(var(--pointer-y) - 210px);background:radial-gradient(circle,rgba(255,252,240,.74),rgba(255,250,239,.12) 37%,transparent 68%);mix-blend-mode:soft-light;transition:left .12s ease-out,top .12s ease-out}.grain{position:absolute;inset:0;opacity:.22;background-image:url("data:image/svg+xml,%3Csvg viewBox='0 0 160 160' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.9' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='.22'/%3E%3C/svg%3E")}
.hero-copy{position:relative;z-index:1}.eyeline{display:flex;align-items:center;gap:9px;margin:0 0 20px;color:#75675d;font-size:13px;font-weight:600}.eyeline i{width:20px;height:1px;background:#cf6c35}.hero h1,.workflow h2,.final h2{margin:0;font:900 clamp(45px,5.35vw,76px)/1.1 'Noto Serif SC',serif;letter-spacing:-.092em}.lede{max-width:555px;margin:24px 0 30px;color:#5c6967;font-size:16px;line-height:1.9}.brief-box{position:relative;display:grid;grid-template-columns:minmax(0,1fr) 48px;gap:8px;max-width:590px;padding:11px 11px 30px;background:#fbfaf3;border:1px solid rgba(36,68,64,.13);color:#142e3d;border-radius:9px;box-shadow:0 17px 36px rgba(93,55,32,.1);transition:transform .25s,box-shadow .25s,border-color .25s}.brief-box.focused{transform:translateY(-3px);border-color:#c96b38;box-shadow:0 23px 47px rgba(93,55,32,.17)}.brief-box textarea{width:100%;resize:none;min-height:45px;padding:4px 5px;border:0;outline:0;color:#15323c;background:transparent;line-height:1.55}.brief-box button{width:44px;height:44px;display:grid;place-items:center;border:0;border-radius:6px;align-self:start;color:#fffdf7;background:#1b5c60;transition:background .2s,transform .2s}.brief-box button:not(:disabled):hover{background:#104448;transform:translateX(2px)}.brief-box button:disabled{color:#81969a;background:#e1dfd5;cursor:not-allowed}.brief-box span{position:absolute;bottom:10px;left:16px;color:#77817d;font-size:11px}.proof{display:flex;gap:16px;flex-wrap:wrap;margin-top:19px;color:#626e6b;font-size:12px}.proof span{display:flex;gap:4px;align-items:center}.proof svg{color:#bf6130}.hero-object{position:relative;justify-self:center;z-index:1;width:min(39vw,500px);transform:translate3d(calc(var(--tilt-x) * .18),calc(var(--tilt-y) * .18),0);transition:transform .5s cubic-bezier(.2,.8,.2,1)}.object-caption{margin:0 0 14px 8%;color:#7f7167;font:500 10px 'DM Mono',monospace;letter-spacing:.12em}.scroll-cue{position:absolute;bottom:27px;left:0;display:flex;align-items:center;gap:10px;padding:0;border:0;color:#6f6259;background:transparent;font-size:12px}.scroll-cue i{display:block;width:33px;height:1px;background:#b96740;position:relative}.scroll-cue i:after{content:'';position:absolute;right:0;top:-3px;width:7px;height:7px;border-right:1px solid #b96740;border-bottom:1px solid #b96740;transform:rotate(45deg)}
.workflow{width:min(1220px,calc(100% - 48px));margin:auto;padding:120px 0}.workflow-heading{max-width:640px}.workflow h2{font-size:clamp(38px,4.6vw,63px)}.sequence{display:grid;grid-template-columns:repeat(3,1fr);gap:22px;margin-top:84px}.sequence article{position:relative;min-height:285px;padding:25px 24px;border-top:1px solid #b9c5bd;transition:border-color .3s,transform .4s}.sequence article:hover{border-color:#d06c35;transform:translateY(-5px)}.sequence b{display:block;margin-bottom:67px;color:#c66432;font:500 15px 'DM Mono',monospace;letter-spacing:.05em}.sequence h3{margin:0 0 12px;font:700 26px 'Noto Serif SC',serif;letter-spacing:-.06em}.sequence p{max-width:290px;margin:0;color:#64716e;font-size:14px;line-height:1.8}.sequence-mark{position:absolute;right:22px;bottom:20px;color:#a2aaa2;font:500 10px 'DM Mono',monospace;letter-spacing:.12em;writing-mode:vertical-rl}.final{display:flex;justify-content:space-between;align-items:end;gap:30px;padding:100px max(24px,calc((100% - 1220px)/2)) 110px;background:#1b575a;color:#f8f3e8}.final .eyeline{color:#c1d7ce}.final .eyeline i{background:#d48450}.final h2{font-size:clamp(39px,4.8vw,65px)}.final-cta{padding:15px 19px;color:#173e43;background:#f7e4d2;white-space:nowrap}.final-cta:hover{color:#fff9ef;background:#d56d2d}.sr-only{position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap;border:0}
[data-reveal]{opacity:0;transform:translateY(24px);transition:opacity .75s ease,transform .75s cubic-bezier(.2,.7,.2,1)}.is-visible{opacity:1;transform:none}.sequence article:nth-child(2){transition-delay:.09s}.sequence article:nth-child(3){transition-delay:.18s}
@media(max-width:760px){.home-nav,.hero,.workflow{width:min(100% - 32px,1220px)}.home-nav{height:64px}.home-nav nav{gap:14px}.home-nav .text-link:first-child{display:none}.hero{grid-template-columns:1fr;min-height:0;padding:70px 0 92px}.hero-atmosphere{inset:-20px -16px 0}.hero h1{font-size:50px}.lede{font-size:15px}.hero-object{width:min(84vw,430px);margin:42px auto 0}.scroll-cue{bottom:28px}.workflow{padding:78px 0}.sequence{grid-template-columns:1fr;gap:0;margin-top:46px}.sequence article{min-height:228px;padding-left:0}.sequence b{margin-bottom:35px}.final{align-items:start;flex-direction:column;padding:70px 16px}.final h2{font-size:45px}}
@media(prefers-reduced-motion:reduce){.mist,.pointer-light,.hero-object,.brief-box,.small-button,.final-cta,[data-reveal],.sequence article{transition:none}.pointer-light{display:none}[data-reveal]{opacity:1;transform:none}}
</style>
