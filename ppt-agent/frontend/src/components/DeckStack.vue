<script setup lang="ts">
const props = withDefaults(defineProps<{ active?: boolean }>(), { active: false })
</script>
<template>
  <div class="deck-stack" :class="{ active: props.active }" aria-hidden="true">
    <div class="paper paper-back"><i></i><i></i><i></i></div>
    <div class="paper paper-mid"><span class="mini-title">研究框架</span><b>从问题<br>到叙事</b><div class="mini-bars"><i></i><i></i><i></i></div></div>
    <div class="paper paper-front"><div class="slide-kicker">DECKFORM / 01</div><h3>让每一页<br>有它的理由。</h3><div class="signal"><span></span>结构正在成稿</div><div class="corner-lines"><i></i><i></i><i></i></div></div>
  </div>
</template>
<style scoped>
.deck-stack { position: relative; width: min(36vw, 470px); aspect-ratio: 1.18; perspective: 1000px; }.paper { position: absolute; inset: 0; padding: 9%; overflow: hidden; color: #102438; background: #eff1e8; border: 1px solid rgba(14,41,53,.13); box-shadow: 0 26px 42px rgba(0,0,0,.18); transition: transform .55s cubic-bezier(.2,.8,.2,1); }.paper-back { transform: rotate(12deg) translate(13%, 2%); background: #b8ded8; }.paper-mid { transform: rotate(5deg) translate(7%, 1%); background: #e6e4d6; }.paper-front { transform: rotate(-2deg); background: #f8f8f1; }.deck-stack.active .paper-back { transform: rotate(17deg) translate(18%, 3%); }.deck-stack.active .paper-mid { transform: rotate(7deg) translate(9%, 0); }.mini-title,.slide-kicker { display: block; font: 500 10px 'DM Mono', monospace; letter-spacing: .1em; color: #427877; }.paper-mid b { display: block; margin-top: 24%; font: 700 clamp(22px,3vw,39px)/1.18 'Noto Serif SC', serif; }.mini-bars { position: absolute; left: 9%; right: 9%; bottom: 12%; display: grid; gap: 7px; }.mini-bars i { display:block; height: 9px; background: #2b5b63; }.mini-bars i:nth-child(2) { width: 74%; background:#69b5aa }.mini-bars i:nth-child(3) { width: 50%; background:#a0c7bd }.paper-front h3 { margin: 22% 0 0; font: 900 clamp(26px,4vw,54px)/1.18 'Noto Serif SC', serif; letter-spacing: -.07em; }.signal { position: absolute; bottom: 10%; left: 9%; display: flex; gap: 7px; align-items: center; padding: 7px 9px; font-size: 11px; color:#166b63; background:#d6f0e7; border-radius: 5px; }.signal span { width:7px;height:7px;border-radius:50%;background:#15a993; box-shadow: 0 0 0 5px rgba(21,169,147,.12); }.corner-lines { position:absolute; right:9%;bottom:10%; display:grid; gap:5px; }.corner-lines i { display:block; width:36px;height:3px;background:#1c4f5c; }.corner-lines i:nth-child(2){width:25px}.corner-lines i:nth-child(3){width:16px}
@media (prefers-reduced-motion: reduce) { .paper { transition: none; } }
</style>
