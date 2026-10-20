<script lang="ts">
  import { onMount } from 'svelte'
  import { Activity, Coins, Cpu, HardDrive, KeyRound, Layers3, Network, Server } from 'lucide-svelte'
  import { api, type APIKey, type Combo, type ProviderConnection, type SystemMetrics } from '../api/client'
  import { language } from '../lib/locale.svelte'

  let { connections = [], combos = [], apiKeys = [] }: {
    connections?: ProviderConnection[]
    combos?: Combo[]
    apiKeys?: APIKey[]
  } = $props()
  let metrics = $state<SystemMetrics | null>(null)
  let usage = $state<any>(null)
  let loadError = $state(false)

  async function refresh() {
    try {
      const [m, u] = await Promise.all([api.getSystemMetrics(), api.getUsageStats('today')])
      metrics = m
      usage = u
      loadError = false
    } catch { loadError = true }
  }
  onMount(() => {
    void refresh()
    const timer = setInterval(() => { if (!document.hidden) void refresh() }, 5000)
    return () => clearInterval(timer)
  })

  const number = (n: number | undefined) => (n ?? 0).toLocaleString(language.current === 'pt-BR' ? 'pt-BR' : 'en-US')
  const bytes = (n: number | undefined) => {
    if (n === undefined || !Number.isFinite(n)) return '—'
    if (!n) return '0 MB'
    const mb = n / 1024 / 1024
    return mb >= 1024 ? `${(mb / 1024).toFixed(2)} GB` : `${mb.toFixed(0)} MB`
  }
  let labels = $derived(language.current === 'pt-BR'
    ? { title: 'Visão geral', subtitle: 'Saúde do serviço, da VPS e atividade recente.', service: 'Serviço 9router-go', vps: 'Servidor VPS', disk: 'Disco da VPS (/)', diskTotal: 'total', cpu: 'CPU', memory: 'Memória', diskUsed: 'Usado', diskAvailable: 'Disponível', providers: 'Provedores ativos', combos: 'Combos', keys: 'Chaves de API', requests: 'Requisições hoje', tokens: 'Tokens hoje', unavailable: 'Métricas indisponíveis no momento' }
    : { title: 'Dashboard', subtitle: 'Service health, VPS resources, and recent activity.', service: '9router-go service', vps: 'VPS server', disk: 'VPS disk (/)', diskTotal: 'total', cpu: 'CPU', memory: 'Memory', diskUsed: 'Used', diskAvailable: 'Available', providers: 'Active providers', combos: 'Combos', keys: 'API keys', requests: 'Requests today', tokens: 'Tokens today', unavailable: 'Metrics are currently unavailable' })
</script>

<section class="space-y-7">
  <div>
    <h2 class="text-2xl font-bold text-text-main">{labels.title}</h2>
    <p class="mt-1 text-sm text-text-muted">{labels.subtitle}</p>
  </div>

  {#if loadError && !metrics}<div class="rounded-xl border border-amber-500/30 bg-amber-500/10 p-4 text-sm text-amber-700">{labels.unavailable}</div>{/if}

  <div class="grid gap-5 xl:grid-cols-3">
    {#each [{ title: labels.service, cpu: metrics?.processCpuPercent, memory: metrics?.processMemoryBytes, percent: metrics?.processMemoryPercent, icon: Cpu }, { title: labels.vps, cpu: metrics?.systemCpuPercent, memory: metrics?.systemMemoryUsedBytes, percent: metrics?.systemMemoryPercent, icon: Server }] as item}
      <article class="rounded-2xl border border-border-subtle bg-surface p-5 shadow-sm">
        <div class="mb-5 flex items-center gap-2 font-semibold text-text-main"><item.icon class="h-5 w-5 text-brand-500" />{item.title}</div>
        <div class="grid grid-cols-2 gap-4">
          <div class="rounded-xl bg-surface-2 p-4"><div class="flex items-center gap-2 text-xs text-text-muted"><Activity class="h-4 w-4" />{labels.cpu}</div><div class="mt-2 text-2xl font-bold text-text-main">{item.cpu?.toFixed(1) ?? '—'}%</div><div class="mt-2 h-1.5 overflow-hidden rounded-full bg-border"><div class="h-full rounded-full bg-gradient-to-r from-emerald-500 to-yellow-400" style={`width:${Math.min(item.cpu ?? 0, 100)}%`}></div></div></div>
          <div class="rounded-xl bg-surface-2 p-4"><div class="flex items-center gap-2 text-xs text-text-muted"><HardDrive class="h-4 w-4" />{labels.memory}</div><div class="mt-2 text-2xl font-bold text-text-main">{bytes(item.memory)}</div><div class="mt-1 text-xs text-text-muted">{item.percent?.toFixed(1) ?? '—'}%</div><div class="mt-2 h-1.5 overflow-hidden rounded-full bg-border"><div class="h-full rounded-full bg-gradient-to-r from-sky-500 to-violet-500" style={`width:${Math.min(item.percent ?? 0, 100)}%`}></div></div></div>
        </div>
      </article>
    {/each}
    <article class="rounded-2xl border border-border-subtle bg-surface p-5 shadow-sm">
      <div class="mb-5 flex items-center gap-2 font-semibold text-text-main"><HardDrive class="h-5 w-5 text-orange-500" />{labels.disk}</div>
      <div class="flex items-baseline justify-between gap-3">
        <div class="text-2xl font-bold text-text-main">{bytes(metrics?.systemDiskUsedBytes)}</div>
        <div class="text-sm text-text-muted">{bytes(metrics?.systemDiskBytes)} {labels.diskTotal}</div>
      </div>
      <div class="mt-3 h-2 overflow-hidden rounded-full bg-border"><div class="h-full rounded-full bg-gradient-to-r from-emerald-500 via-yellow-400 to-orange-500" style={`width:${Math.min(metrics?.systemDiskPercent ?? 0, 100)}%`}></div></div>
      <div class="mt-3 flex justify-between text-xs text-text-muted"><span>{labels.diskUsed}: {metrics?.systemDiskPercent?.toFixed(1) ?? '—'}%</span><span>{labels.diskAvailable}: {bytes(metrics?.systemDiskAvailableBytes)}</span></div>
    </article>
  </div>

  <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
    {#each [
      { label: labels.providers, value: `${connections.filter((c) => c.isActive === 1).length}/${connections.length}`, icon: Network, tint: 'text-sky-500', bg: 'bg-sky-500/10' },
      { label: labels.combos, value: number(combos.length), icon: Layers3, tint: 'text-violet-500', bg: 'bg-violet-500/10' },
      { label: labels.keys, value: number(apiKeys.length), icon: KeyRound, tint: 'text-amber-500', bg: 'bg-amber-500/10' },
      { label: labels.requests, value: number(usage?.totalRequests ?? usage?.requests), icon: Activity, tint: 'text-emerald-500', bg: 'bg-emerald-500/10' },
      { label: labels.tokens, value: number((usage?.totalPromptTokens ?? 0) + (usage?.totalCompletionTokens ?? 0)), icon: Coins, tint: 'text-pink-500', bg: 'bg-pink-500/10' },
    ] as card}
      <article class="flex items-center gap-4 rounded-2xl border border-border-subtle bg-surface p-5 shadow-sm"><div class={`rounded-xl p-3 ${card.bg}`}><card.icon class={`h-5 w-5 ${card.tint}`} /></div><div><p class="text-xs text-text-muted">{card.label}</p><p class="mt-1 text-xl font-bold text-text-main">{card.value}</p></div></article>
    {/each}
  </div>
</section>
