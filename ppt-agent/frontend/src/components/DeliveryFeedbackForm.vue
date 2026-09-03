<script setup lang="ts">
import { ref, watch } from 'vue'
import { Check, Star } from 'lucide-vue-next'
import { saveTaskFeedback } from '../api'
import type { DeliveryFeedback, TaskInfo } from '../types'

const props = defineProps<{ taskId: string; feedback?: DeliveryFeedback }>()
const emit = defineEmits<{ saved: [task: TaskInfo] }>()
const rating = ref(props.feedback?.rating || 0); const suggestion = ref(props.feedback?.suggestion || ''); const busy = ref(false); const error = ref(''); const done = ref(false)
watch(() => props.feedback, feedback => { rating.value = feedback?.rating || 0; suggestion.value = feedback?.suggestion || ''; done.value = Boolean(feedback) })
async function save() { if (!rating.value) { error.value = '请选择 1–5 分'; return }; busy.value = true; error.value = ''; try { const task = await saveTaskFeedback(props.taskId, rating.value, suggestion.value); done.value = true; emit('saved', task) } catch (cause) { error.value = cause instanceof Error ? cause.message : '评分暂时无法保存' } finally { busy.value = false } }
</script>
<template><section class="feedback"><div class="feedback-title"><span>{{ done ? '已记录你的评价' : '这份演示是否符合预期？' }}</span><small>评分不会触发重新生成</small></div><div class="stars" role="radiogroup" aria-label="演示评分"><button v-for="value in 5" :key="value" type="button" :class="{on:value<=rating}" :aria-label="`${value} 分`" @click="rating=value"><Star :size="19" :fill="value<=rating ? 'currentColor' : 'none'" /></button></div><textarea v-model="suggestion" rows="2" maxlength="1000" placeholder="可选：哪一页最有用，或下次希望怎么改？" /><footer><span v-if="error" class="error">{{ error }}</span><span v-else-if="done" class="saved"><Check :size="14" />已保存</span><button :disabled="busy || !rating" @click="save">{{ busy ? '保存中…' : done ? '更新评价' : '提交评价' }}</button></footer></section></template>
<style scoped>
.feedback{max-width:760px;margin:0 auto 22px;padding:15px 17px;border:1px solid #c8dcd5;border-radius:9px;color:#29464d;background:#e6f2ed}.feedback-title{display:flex;justify-content:space-between;gap:12px;align-items:baseline}.feedback-title span{font:700 15px 'Noto Serif SC',serif}.feedback-title small{color:#668087;font-size:11px}.stars{display:flex;gap:4px;margin:10px 0}.stars button{display:grid;place-items:center;padding:3px;border:0;color:#a9beb9;background:transparent}.stars button.on{color:#118e7e}.feedback textarea{width:100%;resize:vertical;padding:8px 9px;border:1px solid #c3d5cf;border-radius:5px;outline:0;color:#244047;background:#f9fcf8;font-size:12px;line-height:1.5}.feedback footer{display:flex;align-items:center;justify-content:space-between;gap:8px;margin-top:9px;font-size:12px}.feedback footer button{border:0;border-radius:5px;padding:7px 10px;color:#edfaf5;background:#0f5961;font-weight:700}.feedback footer button:disabled{opacity:.45}.error{color:#ad3a39}.saved{display:flex;align-items:center;gap:4px;color:#1b776e}
</style>
