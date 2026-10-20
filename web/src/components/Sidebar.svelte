<script lang="ts">
  import { api, type SystemVersionInfo } from '../api/client'
  import { copyToClipboard } from '../lib/clipboard'
  import { TAB_ROUTES, type ActiveTab } from '../lib/router'
  import { isNavLinkActive, visibleNavGroups } from '../lib/nav'
  import { language } from '../lib/locale.svelte'

  export type { ActiveTab }
  let {
    activeTab = $bindable('dashboard'),
    navigate = (tab: ActiveTab) => {
      activeTab = tab
    },
    activeConnections = 0,
    totalConnections = 0,
    onClose,
    showUpdateModal = $bindable(false),
    updateInfo = $bindable(null),
    version = $bindable(''),
    hiddenNav,
    brandName = '',
    brandLogo = '',
  }: {
    activeTab: ActiveTab
    navigate?: (tab: ActiveTab, replace?: boolean) => void
    activeConnections?: number
    totalConnections?: number
    onClose?: () => void
    showUpdateModal?: boolean
    updateInfo?: SystemVersionInfo | null
    version?: string
    hiddenNav?: string[]
    brandName?: string
    brandLogo?: string
  } = $props()
  const INSTALL_CMD = '9router-go update'
  let copied = $state(false)

  const navIconTone: Record<string, string> = {
    dashboard: 'text-cyan-500', endpoint: 'text-blue-500', connections: 'text-sky-500',
    combos: 'text-violet-500', analytics: 'text-emerald-500', quota: 'text-amber-500',
    'token-saver': 'text-pink-500', 'cli-tools': 'text-indigo-500',
    'media-providers': 'text-fuchsia-500', 'media-embedding': 'text-teal-500',
    'media-image': 'text-rose-500', 'media-tts': 'text-orange-500', 'media-stt': 'text-amber-600',
    'media-video': 'text-red-500', 'media-systemone': 'text-purple-500', 'media-web': 'text-blue-400',
    'proxy-pools': 'text-cyan-600', skills: 'text-lime-600', 'console-log': 'text-slate-500',
    settings: 'text-slate-500', security: 'text-emerald-600', plans: 'text-yellow-500',
  }

  function iconTone(id: string, active: boolean) {
    return active ? 'text-primary' : `${navIconTone[id] || 'text-text-muted'} group-hover:scale-110 transition-all`
  }

  // Media providers accordion (collapsed by default)
  let isMediaOpen = $state(false)

  $effect(() => {
    api
      .getSystemVersion()
      .then((v) => {
        if (v) {
          version = v.currentVersion || ''
          if (v.hasUpdate) {
            updateInfo = v
          }
        }
      })
      .catch(() => {})

    const timer = setTimeout(() => {
      api
        .checkUpdate()
        .then((v) => {
          if (v) {
            version = v.currentVersion || version
            if (v.hasUpdate) {
              updateInfo = v
            }
          }
        })
        .catch(() => {})
    }, 2500)

    return () => clearTimeout(timer)
  })

  async function copyInstallCmd() {
    await copyToClipboard(INSTALL_CMD)
    copied = true
    setTimeout(() => {
      copied = false
    }, 2000)
  }

  function handleNav(tab: ActiveTab, e?: MouseEvent) {
    if (e) {
      if (e.ctrlKey || e.metaKey || e.shiftKey || e.altKey || (e.button !== undefined && e.button !== 0)) return
      e.preventDefault()
    }
    navigate(tab)
    onClose?.()
  }

  // The sidebar and the Settings > Sidebar switch read the same entry list
  // (web/src/lib/nav.ts), so a hidden entry is hidden in both.
  const navGroups = $derived(visibleNavGroups(hiddenNav).map((group) => ({
    ...group,
    label: language.current === 'pt-BR' ? ({ System: 'Sistema' }[group.label || ''] || group.label) : group.label,
    entries: group.entries.map((entry) => {
      const tr: Record<string, string> = { Dashboard: 'Painel', 'Endpoint & Key': 'Endpoint e chaves', Providers: 'Provedores', 'Combo & Vision Adapter': 'Combos e adaptador de visão', Usage: 'Uso', 'Quota Tracker': 'Cotas', 'Token Saver': 'Economia de tokens', 'CLI Tools': 'Ferramentas CLI', 'Media Providers': 'Provedores de mídia', Embedding: 'Embeddings', 'Text to Image': 'Texto para imagem', 'Text To Speech': 'Texto para fala', 'Speech To Text': 'Fala para texto', Video: 'Vídeo', 'System One': 'System One', 'Web Fetch & Search': 'Busca e pesquisa web', 'Proxy Pools': 'Pools de proxy', Skills: 'Skills', 'Console Log': 'Log do console', Settings: 'Configurações', Security: 'Segurança', Plans: 'Planos' }
      if (entry.kind === 'link') return { ...entry, label: language.current === 'pt-BR' ? (tr[entry.label] || entry.label) : entry.label }
      return { ...entry, label: language.current === 'pt-BR' ? (tr[entry.label] || entry.label) : entry.label, children: entry.children.map((child) => ({ ...child, label: language.current === 'pt-BR' ? (tr[child.label] || child.label) : child.label })) }
    }),
  })))
</script>

<aside
  class="flex h-full max-h-screen w-72 flex-col overflow-hidden border-r border-border-subtle bg-sidebar backdrop-blur-xl transition-colors duration-300 flex-shrink-0 select-none z-30"
>
  <!-- Window control / traffic lights -->
  <div class="flex items-center gap-2 px-6 pt-5 pb-2 shrink-0">
    <div class="w-3 h-3 rounded-full bg-[#FF5F56]"></div>
    <div class="w-3 h-3 rounded-full bg-[#FFBD2E]"></div>
    <div class="w-3 h-3 rounded-full bg-[#27C93F]"></div>
  </div>

  <!-- Brand header: 9router-go with official favicon.svg logo -->
  <div class="px-6 py-4 flex flex-col gap-2 shrink-0">
    <a
      href={TAB_ROUTES.dashboard}
      onclick={(e) => handleNav('dashboard', e)}
      class="flex items-center gap-3 cursor-pointer group"
    >
      <div
        class="flex items-center justify-center size-9 rounded-[10px] bg-surface-2 border border-border-subtle shadow-[var(--shadow-warm)] flex-shrink-0 group-hover:scale-105 transition-transform overflow-hidden p-1.5"
      >
        <img
          src={brandLogo || '/favicon.svg'}
          alt={brandName || '9router-go'}
          class="w-full h-full object-contain"
        />
      </div>
      <div class="flex flex-col min-w-0">
        <h1 class="text-lg font-semibold tracking-tight text-text-main truncate leading-snug">
          {brandName || '9router-go'}
        </h1>
        <span class="text-xs text-text-muted leading-tight">
          {version ? `v${version}` : 'v1.9.9'}
        </span>
      </div>
    </a>

    <!-- Update notification banner below version (matches upstream 9router) -->
    {#if updateInfo && updateInfo.hasUpdate}
      <div
        class="flex flex-col gap-1.5 rounded-lg p-2 bg-green-500/10 dark:bg-amber-500/10 border border-green-500/30 dark:border-amber-500/30 text-green-700 dark:text-amber-400 mt-0.5 animate-in fade-in duration-200"
      >
        <span class="text-[11px] font-semibold text-green-700 dark:text-amber-400 flex items-center gap-1">
          <span class="material-symbols-outlined text-[14px]">arrow_upward</span>
          <span class="truncate">New version available: v{updateInfo.latestVersion}</span>
        </span>
        <div class="flex items-center gap-1.5">
          <button
            type="button"
            onclick={() => (showUpdateModal = true)}
            class="px-2 py-0.5 rounded bg-green-600 hover:bg-green-700 dark:bg-amber-500 dark:hover:bg-amber-600 text-white text-[11px] font-semibold transition-colors cursor-pointer shrink-0"
          >
            Update now
          </button>
          <button
            type="button"
            onclick={copyInstallCmd}
            title="Copy install command"
            class="flex-1 text-left hover:opacity-80 transition-opacity cursor-pointer min-w-0"
          >
            <code class="block text-[10px] text-green-700 dark:text-amber-300 font-mono truncate bg-surface/70 px-1 py-0.5 rounded">
              {copied ? '✓ copied!' : INSTALL_CMD}
            </code>
          </button>
        </div>
      </div>
    {/if}
  </div>

  <!-- Navigation -->
  <nav class="flex-1 min-h-0 px-4 py-2 space-y-0.5 overflow-y-auto custom-scrollbar">
    {#each navGroups as group (group.id)}
      <div class="{group.label ? 'pt-3 mt-2' : ''} space-y-0.5">
        {#if group.label}
          <p class="px-3 text-xs font-semibold text-text-muted/60 uppercase tracking-wider mb-2">
            {group.label}
          </p>
        {/if}

        {#each group.entries as entry (entry.id)}
          {#if entry.kind === 'accordion'}
            <button
              type="button"
              onclick={() => (isMediaOpen = !isMediaOpen)}
              class="w-full flex items-center gap-3 px-3 py-1.5 rounded-lg transition-all group cursor-pointer {activeTab.startsWith(
                'media-'
              )
                ? 'bg-primary/10 text-primary font-medium'
                : 'text-text-muted hover:bg-surface-2 hover:text-text-main'}"
            >
              <span class={`material-symbols-outlined text-[18px] ${iconTone(entry.id, activeTab.startsWith('media-'))}`}>{entry.icon}</span>
              <span class="text-[13px] flex-1 text-left">{entry.label}</span>
              <span
                class="material-symbols-outlined text-[14px] transition-transform duration-200"
                style:transform={isMediaOpen ? 'rotate(180deg)' : 'rotate(0deg)'}
              >
                expand_more
              </span>
            </button>

            {#if isMediaOpen}
              <div class="pl-4 space-y-0.5">
                {#each entry.children as child (child.id)}
                  {@const childActive = isNavLinkActive(child.id, activeTab)}
                  <a
                    href={TAB_ROUTES[child.id]}
                    onclick={(e) => handleNav(child.id, e)}
                    class="flex items-center gap-3 px-3 py-1.5 rounded-lg transition-all group cursor-pointer {childActive
                      ? 'bg-primary/10 text-primary font-medium'
                      : 'text-text-muted hover:bg-surface-2 hover:text-text-main'}"
                  >
                    <span
                      class={`material-symbols-outlined text-[16px] ${childActive ? 'fill-1 text-primary' : iconTone(child.id, false)}`}
                    >
                      {child.icon}
                    </span>
                    <span class="text-[13px]">{child.label}</span>
                  </a>
                {/each}
              </div>
            {/if}
          {:else}
            {@const active = isNavLinkActive(entry.id, activeTab)}
            <a
              href={TAB_ROUTES[entry.id]}
              onclick={(e) => handleNav(entry.id, e)}
              class="flex items-center gap-3 px-3 py-1.5 rounded-lg transition-all group cursor-pointer {active
                ? 'bg-primary/10 text-primary font-medium'
                : 'text-text-muted hover:bg-surface-2 hover:text-text-main'}"
            >
              <span
                class={`material-symbols-outlined text-[18px] ${active ? 'fill-1 text-primary' : iconTone(entry.id, false)}`}
              >
                {entry.icon}
              </span>
              <span class="text-[13px]">{entry.label}</span>
            </a>
          {/if}
        {/each}
      </div>
    {/each}
  </nav>

  <!-- Bottom connection summary -->
  <div class="p-4 border-t border-border-subtle shrink-0">
    <div
      class="p-2.5 rounded-[10px] bg-surface border border-border-subtle flex items-center justify-between"
    >
      <div class="min-w-0">
        <p class="text-[11px] text-text-muted uppercase tracking-wide">Providers</p>
        <p class="text-xs font-semibold text-text-main mt-0.5 truncate">
          {activeConnections} <span class="text-text-muted font-normal">/ {totalConnections} active</span>
        </p>
      </div>
      <span class="material-symbols-outlined text-primary text-[18px]">radio_button_checked</span>
    </div>
  </div>
</aside>

