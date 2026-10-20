import type { Locale } from './locale.svelte'
import { generatedPtBR } from './page-translations.generated'

// Translate only exact interface strings. Provider/model names, logs, prompts,
// API responses, and anything typed by the user are left untouched.
const ptBR: Record<string, string> = {
  'Overview': 'Visão geral', 'Dashboard': 'Painel', 'Endpoint & Key': 'Endpoint e chaves',
  'API endpoint and key configuration': 'Configuração de endpoint e chaves de API',
  'Providers': 'Provedores', 'Connections': 'Conexões', 'Manage your AI provider connections': 'Gerencie suas conexões com provedores de IA',
  'Combos': 'Combos', 'Combo & Routing': 'Combos e roteamento', 'Model combos and failover strategies': 'Combos de modelos e estratégias de contingência',
  'Usage': 'Uso', 'Usage & Analytics': 'Uso e análise', 'Analytics': 'Análises', 'Cache': 'Cache', 'Compression': 'Compressão',
  'Quota Tracker': 'Cotas', 'Token Saver': 'Economia de tokens', 'CLI Tools': 'Ferramentas CLI',
  'Proxy Pools': 'Pools de proxy', 'Agent Skills': 'Skills de agentes', 'Console Log': 'Log do console',
  'Settings': 'Configurações', 'Security': 'Segurança', 'Plans': 'Planos', 'Login': 'Entrar',
  'Save': 'Salvar', 'Save Changes': 'Salvar alterações', 'Cancel': 'Cancelar', 'Close': 'Fechar',
  'Back': 'Voltar', 'Next': 'Próximo', 'Previous': 'Anterior', 'Continue': 'Continuar',
  'Confirm': 'Confirmar', 'Delete': 'Excluir', 'Remove': 'Remover', 'Edit': 'Editar', 'Add': 'Adicionar',
  'Create': 'Criar', 'Update': 'Atualizar', 'Refresh': 'Atualizar', 'Search': 'Buscar', 'Filter': 'Filtrar',
  'Clear': 'Limpar', 'Reset': 'Redefinir', 'Copy': 'Copiar', 'Copied!': 'Copiado!', 'Apply': 'Aplicar',
  'Configure': 'Configurar', 'Test': 'Testar', 'Test Connection': 'Testar conexão', 'Loading...': 'Carregando...',
  'Connecting...': 'Conectando...', 'No data available': 'Nenhum dado disponível', 'No results found': 'Nenhum resultado encontrado',
  'No results': 'Nenhum resultado', 'All': 'Todos', 'None': 'Nenhum', 'Optional': 'Opcional', 'Required': 'Obrigatório',
  'Enabled': 'Ativado', 'Disabled': 'Desativado', 'Enable': 'Ativar', 'Disable': 'Desativar',
  'Active': 'Ativo', 'Inactive': 'Inativo', 'Connected': 'Conectado', 'Disconnected': 'Desconectado',
  'Pending': 'Pendente', 'Success': 'Sucesso', 'Failed': 'Falhou', 'Error': 'Erro', 'Status': 'Status',
  'Name': 'Nome', 'Description': 'Descrição', 'Type': 'Tipo', 'Actions': 'Ações', 'Details': 'Detalhes',
  'Provider': 'Provedor', 'Model': 'Modelo', 'Models': 'Modelos', 'Endpoint': 'Endpoint', 'Base URL': 'URL base',
  'API Key': 'Chave de API', 'API Keys': 'Chaves de API', 'Key': 'Chave', 'URL': 'URL', 'Port': 'Porta',
  'Protocol': 'Protocolo', 'Authentication': 'Autenticação', 'Connection': 'Conexão', 'Account': 'Conta',
  'Priority': 'Prioridade', 'Strategy': 'Estratégia', 'Fallback': 'Contingência', 'Round Robin': 'Alternância circular',
  'Random': 'Aleatório', 'Quality': 'Qualidade', 'Latency': 'Latência', 'Requests': 'Requisições', 'Tokens': 'Tokens',
  'Today': 'Hoje', 'Yesterday': 'Ontem', 'Last 7 days': 'Últimos 7 dias', 'Last 30 days': 'Últimos 30 dias',
  'Custom': 'Personalizado', 'Date': 'Data', 'Time': 'Horário', 'Duration': 'Duração', 'Average': 'Média',
  'Total': 'Total', 'Used': 'Usado', 'Available': 'Disponível', 'Remaining': 'Restante', 'Limit': 'Limite',
  'Usage could not be loaded. The figures below are from the last successful read.': 'Não foi possível carregar o uso. Os valores abaixo são da última consulta bem-sucedida.',
  'No activity in the selected period': 'Nenhuma atividade no período selecionado',
  'Recent Requests': 'Requisições recentes', 'Avg Latency': 'Latência média', 'Avg Speed': 'Velocidade média',
  'Avg TTFT': 'TTFT médio', 'Last Used': 'Último uso', 'Health': 'Saúde', 'Success Rate': 'Taxa de sucesso',
  'Request Details': 'Detalhes da requisição', 'Request History': 'Histórico de requisições',
  'Failed to load request history': 'Falha ao carregar o histórico de requisições',
  'Input Tokens': 'Tokens de entrada', 'Output Tokens': 'Tokens de saída', 'Prompt Tokens': 'Tokens do prompt',
  'Completion Tokens': 'Tokens de conclusão', 'Cost': 'Custo', 'Estimated Cost': 'Custo estimado',
  'Cache Rate': 'Taxa de cache', 'Cached Requests': 'Requisições em cache', 'Cached Tokens': 'Tokens em cache',
  'Tokens Saved': 'Tokens economizados', 'Cost Saved': 'Custo economizado', 'Hit Rate': 'Taxa de acerto',
  'Hits / Misses': 'Acertos / falhas', 'Memory Entries': 'Entradas em memória', 'Clear Cache': 'Limpar cache',
  'Invalidate cached responses by model or clear the entire store.': 'Invalide respostas em cache por modelo ou limpe todo o armazenamento.',
  'Select a provider': 'Selecione um provedor', 'Select a model': 'Selecione um modelo',
  'Select models': 'Selecione os modelos', 'Add Provider': 'Adicionar provedor', 'Add Connection': 'Adicionar conexão',
  'Add Model': 'Adicionar modelo', 'Add API Key': 'Adicionar chave de API', 'Create API Key': 'Criar chave de API',
  'No providers configured': 'Nenhum provedor configurado', 'No models found': 'Nenhum modelo encontrado',
  'Connect a provider to get started.': 'Conecte um provedor para começar.',
  'Are you sure?': 'Tem certeza?', 'This action cannot be undone.': 'Esta ação não pode ser desfeita.',
  'Copied to clipboard': 'Copiado para a área de transferência', 'Failed to copy': 'Falha ao copiar',
  'Something went wrong': 'Algo deu errado', 'Please try again.': 'Tente novamente.',
  'No API key created': 'Nenhuma chave de API criada', 'Create a key to use this endpoint.': 'Crie uma chave para usar este endpoint.',
  'Default Model': 'Modelo padrão', 'Default model': 'Modelo padrão', 'Current Model': 'Modelo atual',
  'Select a combo': 'Selecione um combo', 'Create Combo': 'Criar combo', 'Edit Combo': 'Editar combo',
  'Add Model to Combo': 'Adicionar modelo ao combo', 'Remove Model': 'Remover modelo',
  'No combos configured': 'Nenhum combo configurado', 'Choose a routing strategy': 'Escolha uma estratégia de roteamento',
  'Failover': 'Alternância por falha', 'Round-robin': 'Alternância circular',
  'Primary': 'Principal', 'Secondary': 'Secundário', 'Fallback Model': 'Modelo de contingência',
  'Daily Limit': 'Limite diário', 'Monthly Limit': 'Limite mensal', 'Rate Limit': 'Limite de requisições',
  'Requests per minute': 'Requisições por minuto', 'Tokens per minute': 'Tokens por minuto',
  'No quota data available': 'Nenhum dado de cota disponível', 'Reset quota': 'Redefinir cota',
  'Never': 'Nunca', 'Hourly': 'A cada hora', 'Daily': 'Diário', 'Weekly': 'Semanal', 'Monthly': 'Mensal',
  'Install Guide': 'Guia de instalação', 'Manual Configuration': 'Configuração manual', 'Manual Config': 'Configuração manual',
  'Installation': 'Instalação', 'Not installed': 'Não instalado', 'Not detected': 'Não detectado',
  'Scan Installed Tools': 'Detectar ferramentas instaladas', 'One-Click Environment Setup': 'Configuração de ambiente em um clique',
  'Copy Export Commands': 'Copiar comandos de configuração', 'Active Token': 'Token ativo', 'Apply to All Profiles': 'Aplicar a todos os perfis',
  'No key created': 'Nenhuma chave criada', 'Current': 'Atual', 'Profile': 'Perfil',
  'Write these into the configuration file when you prefer to edit the files yourself.': 'Grave estas opções no arquivo de configuração se preferir editar os arquivos manualmente.',
  'Tunnel': 'Túnel', 'Not connected': 'Não conectado', 'Enable Tunnel': 'Ativar túnel', 'Disable Tunnel': 'Desativar túnel',
  'Cloudflare Quick Tunnel': 'Túnel rápido do Cloudflare', 'Allow dashboard access via tunnel': 'Permitir acesso ao painel pelo túnel',
  'Pool Name': 'Nome do pool', 'Add Pool': 'Adicionar pool', 'Edit Pool': 'Editar pool', 'Delete Pool': 'Excluir pool',
  'No proxy pools configured': 'Nenhum pool de proxy configurado', 'Proxy URL': 'URL do proxy',
  'Connect your agent': 'Conecte seu agente', 'Copy link': 'Copiar link', 'Open Documentation': 'Abrir documentação',
  'Clear logs': 'Limpar logs', 'Download logs': 'Baixar logs', 'No logs yet': 'Ainda não há logs',
  'Live output': 'Saída em tempo real', 'Terminal': 'Terminal', 'Command': 'Comando', 'Run': 'Executar',
  'Password': 'Senha', 'Enter your password': 'Digite sua senha', 'Sign in': 'Entrar', 'Sign out': 'Sair',
  'Forgot password?': 'Esqueceu a senha?', 'Welcome back': 'Bem-vindo de volta', 'Invalid password': 'Senha inválida',
  'Change password': 'Alterar senha', 'Current password': 'Senha atual', 'New password': 'Nova senha',
  'Confirm password': 'Confirme a senha', 'Save Settings': 'Salvar configurações', 'General': 'Geral',
  'Appearance': 'Aparência', 'Preferences': 'Preferências', 'System': 'Sistema', 'Branding': 'Identidade visual',
  'Application name': 'Nome da aplicação', 'Logo URL': 'URL da logo', 'Preview': 'Prévia',
  'White-label branding': 'Identidade visual personalizada', 'Leave blank to use the default.': 'Deixe em branco para usar o padrão.',
  'Security Settings': 'Configurações de segurança', 'Two-Factor Authentication': 'Autenticação em dois fatores',
  'API Access': 'Acesso à API', 'Allowed IP Addresses': 'Endereços IP permitidos', 'Session Timeout': 'Tempo limite da sessão',
  'No plans available': 'Nenhum plano disponível', 'Create Plan': 'Criar plano', 'Edit Plan': 'Editar plano',
  'Plan Name': 'Nome do plano', 'Price': 'Preço', 'Billing Period': 'Período de cobrança', 'Features': 'Recursos',
  'No providers available': 'Nenhum provedor disponível', 'Manage providers': 'Gerenciar provedores',
  'Image Generation': 'Geração de imagens', 'Text to Speech': 'Texto para fala', 'Speech to Text': 'Fala para texto',
  'Video Generation': 'Geração de vídeo', 'Web Search': 'Busca na web', 'Embeddings': 'Embeddings',
  'No media providers configured': 'Nenhum provedor de mídia configurado', 'Test Provider': 'Testar provedor',
  'Save and Test': 'Salvar e testar', 'Add Endpoint': 'Adicionar endpoint', 'Custom Endpoint': 'Endpoint personalizado',
  'API Key / Token': 'Chave de API / token', 'Display Name': 'Nome de exibição',
  'No activity yet': 'Ainda não há atividade', 'View Details': 'Ver detalhes', 'Show More': 'Mostrar mais',
  'Show Less': 'Mostrar menos', 'Expand': 'Expandir', 'Collapse': 'Recolher', 'Documentation': 'Documentação',
  'Change Log': 'Registro de alterações', 'View on GitHub': 'Ver no GitHub', 'Loading change log...': 'Carregando registro de alterações...',
  'Failed to load change log': 'Falha ao carregar o registro de alterações',
  'Prompt caching and semantic deduplication tracking across all AI providers.': 'Cache de prompts e acompanhamento da desduplicação semântica em todos os provedores de IA.',
  '2. Caveman (Linguistic Economy)': '2. Caveman (economia linguística)',
  '3. ADHD (Action-First Format)': '3. ADHD (formato focado em ações)',
  'Direct, secure connection points to your local AI router': 'Pontos de conexão diretos e seguros para seu roteador de IA local',
  'Tunnel reconnecting...': 'Reconectando ao túnel...', 'Tunnel checking...': 'Verificando o túnel...',
  'Creating tunnel...': 'Criando túnel...', 'Tailscale reconnecting...': 'Reconectando ao Tailscale...',
  'Tailscale checking...': 'Verificando o Tailscale...', 'Connecting to 9router-go Localhost Gateway (:20130)...': 'Conectando ao gateway local do 9router-go (:20130)...',
  'Creating...': 'Criando...', 'API Key Created': 'Chave de API criada',
  'Save this key now!': 'Salve esta chave agora!', 'Enable Tailscale': 'Ativar Tailscale',
  'Disable Tailscale': 'Desativar Tailscale', 'Tailscale Funnel': 'Túnel Funnel do Tailscale',
  'All providers': 'Todos os provedores', 'Hidden:': 'Oculto:', 'Redeem a reset credit': 'Resgatar crédito de redefinição',
  'Redeem this reset credit?': 'Resgatar este crédito de redefinição?',
  'Loading reset credits…': 'Carregando créditos de redefinição…',
  'No reset credits are available on this account.': 'Não há créditos de redefinição disponíveis nesta conta.',
  'No expiry date': 'Sem data de validade', 'recommended': 'recomendado',
  'Strict Proxy': 'Proxy estrito', 'Inactive pools are ignored by runtime resolution.': 'Pools inativos são ignorados durante a resolução das requisições.',
  'Filter:': 'Filtro:', 'Sort:': 'Ordenar:', 'Default': 'Padrão', 'Fastest (Latency asc)': 'Mais rápido (menor latência)',
  'Recently Tested': 'Testados recentemente', 'Name (A-Z)': 'Nome (A-Z)', 'Total:': 'Total:', 'Active:': 'Ativos:', 'Failed:': 'Falhas:',
  'No proxy pool entries yet': 'Ainda não há entradas nos pools de proxy',
  'No proxy pools match current filter': 'Nenhum pool de proxy corresponde ao filtro atual',
  'Active Token:': 'Token ativo:', 'none created': 'nenhum criado',
  'Setup Instructions': 'Instruções de configuração', 'Checking Hermes Agent…': 'Verificando o Hermes Agent…',
  'Hermes Agent not detected locally': 'Hermes Agent não detectado localmente',
  'Hermes Agent — Manual Configuration': 'Hermes Agent — configuração manual',
  'Empty roles inherit the default model.': 'Funções vazias herdam o modelo padrão.',
  'No nodes with 50+ requests in this period': 'Nenhum nó com mais de 50 requisições neste período',
  'No models with 10+ requests under this node': 'Nenhum modelo com mais de 10 requisições neste nó',
  'Custom window': 'Período personalizado', 'Days or hours, for example 14d or 12h.': 'Dias ou horas, por exemplo, 14d ou 12h.',
  'Last 7 days (≥50 real requests)': 'Últimos 7 dias (≥50 requisições reais)',
  'Cached / Total Req': 'Requisições em cache / total', 'Cached Req:': 'Req. em cache:', 'Total Req:': 'Req. totais:',
  'Cached vs input tokens': 'Tokens em cache vs. tokens de entrada', 'Exact & semantic matches': 'Correspondências exatas e semânticas',
  'From cached responses': 'De respostas em cache', 'In-memory LRU store': 'Armazenamento LRU em memória',
  'Read from prompt cache': 'Lido do cache de prompt', 'Written to prompt cache': 'Gravado no cache de prompt',
  'Requests and cached tokens over time.': 'Requisições e tokens em cache ao longo do tempo.',
  'Signature / Key': 'Assinatura / chave', 'Creation Tokens': 'Tokens de criação',
  'Configure your AI provider': 'Configure seu provedor de IA', 'Provider Name': 'Nome do provedor',
  'Provider URL': 'URL do provedor', 'API Base URL': 'URL base da API', 'Organization ID': 'ID da organização',
  'Authentication Type': 'Tipo de autenticação', 'OAuth Authentication': 'Autenticação OAuth',
  'Sign in with': 'Entrar com', 'Connect': 'Conectar', 'Disconnect': 'Desconectar', 'Reconnect': 'Reconectar',
  'Connection successful': 'Conexão realizada com sucesso', 'Connection failed': 'Falha na conexão',
  'Test successful': 'Teste realizado com sucesso', 'Test failed': 'Falha no teste', 'Checking...': 'Verificando...',
  'Sync Models': 'Sincronizar modelos', 'Refresh Models': 'Atualizar modelos', 'Fetch Models': 'Buscar modelos',
  'Last synced': 'Última sincronização', 'Sync now': 'Sincronizar agora', 'No models available': 'Nenhum modelo disponível',
  'Custom model': 'Modelo personalizado', 'Add custom model': 'Adicionar modelo personalizado',
  'Display Name (optional)': 'Nome de exibição (opcional)', 'Model ID': 'ID do modelo',
  'Context Length': 'Tamanho do contexto', 'Max Output Tokens': 'Máximo de tokens de saída',
  'Capabilities': 'Recursos', 'Vision': 'Visão', 'Reasoning': 'Raciocínio', 'Tool Calling': 'Chamada de ferramentas',
  'Model capabilities': 'Recursos do modelo', 'No provider selected': 'Nenhum provedor selecionado',
  'No connections yet': 'Ainda não há conexões', 'Add your first provider': 'Adicione seu primeiro provedor',
  'Search providers...': 'Buscar provedores...', 'Search models...': 'Buscar modelos...',
  'Filter by provider': 'Filtrar por provedor', 'Show inactive': 'Mostrar inativos', 'Show all': 'Mostrar todos',
  'Collapse all': 'Recolher tudo', 'Expand all': 'Expandir tudo', 'Show details': 'Mostrar detalhes',
  'Hide details': 'Ocultar detalhes', 'View all': 'Ver tudo', 'Load more': 'Carregar mais',
  'No requests yet': 'Ainda não há requisições', 'No usage data yet': 'Ainda não há dados de uso',
  'No data for this period': 'Não há dados para este período', 'Select a time period': 'Selecione um período',
  'Start Date': 'Data inicial', 'End Date': 'Data final', 'Apply filter': 'Aplicar filtro',
  'Clear filters': 'Limpar filtros', 'Export CSV': 'Exportar CSV', 'Download CSV': 'Baixar CSV',
  'View raw data': 'Ver dados brutos', 'Request ID': 'ID da requisição', 'Response': 'Resposta',
  'Input': 'Entrada', 'Output': 'Saída', 'Started At': 'Iniciada em', 'Ended At': 'Finalizada em',
  'Cache Source': 'Origem do cache', 'Cached': 'Em cache', 'Completion': 'Conclusão',
  'No API key selected': 'Nenhuma chave de API selecionada', 'Select an API key': 'Selecione uma chave de API',
  'Permissions': 'Permissões', 'Global access': 'Acesso global', 'Allowed Combos': 'Combos permitidos',
  'Unlimited': 'Ilimitado', 'Daily token limit': 'Limite diário de tokens', 'Expires At': 'Expira em',
  'Never expires': 'Nunca expira', 'Key Name': 'Nome da chave', 'Key created successfully': 'Chave criada com sucesso',
  'Key deleted successfully': 'Chave excluída com sucesso', 'Revoke Key': 'Revogar chave', 'Last Used At': 'Último uso em',
  'Install the CLI to use this feature.': 'Instale a CLI para usar este recurso.',
  'Not installed locally': 'Não instalado localmente', 'Detected': 'Detectado', 'Not detected locally': 'Não detectado localmente',
  'Install with': 'Instale com', 'Open install guide': 'Abrir guia de instalação',
  'Environment Variables': 'Variáveis de ambiente', 'Configuration File': 'Arquivo de configuração',
  'Copy Configuration': 'Copiar configuração', 'Apply Configuration': 'Aplicar configuração',
  'Configuration saved': 'Configuração salva', 'Configuration applied': 'Configuração aplicada',
  'No proxy pools found': 'Nenhum pool de proxy encontrado', 'Add Proxy Pool': 'Adicionar pool de proxy',
  'Proxy Pool': 'Pool de proxy', 'Rotation Strategy': 'Estratégia de rotação', 'Test Proxy': 'Testar proxy',
  'Test All Proxies': 'Testar todos os proxies', 'Proxy Status': 'Status do proxy', 'Latency (ms)': 'Latência (ms)',
  'Last Checked': 'Última verificação', 'Strict mode': 'Modo estrito', 'Use proxy for all requests': 'Usar proxy em todas as requisições',
  'Pool is active': 'Pool ativo', 'Pool is inactive': 'Pool inativo', 'Add proxy': 'Adicionar proxy',
  'Proxy URL (http://...)': 'URL do proxy (http://...)', 'Username': 'Usuário',
  'No skills found': 'Nenhuma skill encontrada', 'Copy Skill Link': 'Copiar link da skill',
  'Skill copied to clipboard': 'Link da skill copiado', 'Search skills...': 'Buscar skills...',
  'Live Logs': 'Logs ao vivo', 'Clear Log': 'Limpar log', 'Copy Logs': 'Copiar logs',
  'Scroll to bottom': 'Rolar até o final', 'Pause output': 'Pausar saída', 'Resume output': 'Retomar saída',
  'No output yet': 'Ainda não há saída', 'Connected to server': 'Conectado ao servidor',
  'Connection lost': 'Conexão perdida', 'Reconnecting...': 'Reconectando...',
  'Set password': 'Definir senha', 'Setting password...': 'Definindo senha...', 'Logging in...': 'Entrando...',
  'Wait': 'Aguarde', 'Locked. Retry in': 'Bloqueado. Tente novamente em', 'Authentication required': 'Autenticação necessária',
  'Set up your admin password': 'Configure sua senha de administrador', 'Create a password to secure your dashboard.': 'Crie uma senha para proteger seu painel.',
  'Sign in to your dashboard': 'Entre no seu painel', 'Enter your admin password to continue.': 'Digite sua senha de administrador para continuar.',
  'Observability Mode': 'Modo de observabilidade', 'Diagnostics Retention': 'Retenção de diagnósticos',
  'Max Request Records': 'Máximo de registros de requisições', 'Usage Ledger Retention': 'Retenção do histórico de uso',
  'Feature Visibility': 'Visibilidade dos recursos', 'Show or hide sections in the navigation menu.': 'Mostre ou oculte seções no menu de navegação.',
  'System Resource Usage': 'Uso de recursos do sistema', 'Service CPU': 'CPU do serviço', 'System CPU': 'CPU do sistema',
  'Memory Usage': 'Uso de memória', 'Disk Usage': 'Uso de disco', 'Available Space': 'Espaço disponível',
  'Update available': 'Atualização disponível', 'Check for updates': 'Verificar atualizações',
  'Restart required': 'É necessário reiniciar', 'Restart service': 'Reiniciar serviço',
}

const pageTranslations = { ...generatedPtBR, ...ptBR }
const enByPt = new Map(Object.entries(pageTranslations).map(([en, pt]) => [pt, en]))
type NodeTranslation = { english: string; rendered: string }
const nodeTranslations = new WeakMap<Text, NodeTranslation>()
const attributeTranslations = new WeakMap<Element, Map<string, NodeTranslation>>()
const translatedAttributes = ['title', 'aria-label', 'placeholder', 'aria-description']

function normalize(value: string) {
  return value.trim().replace(/\s+/g, ' ')
}

function translateValue(value: string, locale: Locale) {
  const key = normalize(value)
  if (/-----BEGIN |-----END |^\s*(export|npm install|curl|git clone|go run)\b/i.test(key)) return null
  const english = pageTranslations[key] ? key : enByPt.get(key)
  if (!english) return null
  return locale === 'pt-BR' ? pageTranslations[english] : english
}

function localizeTree(root: HTMLElement, locale: Locale) {
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT)
  let node: Node | null
  while ((node = walker.nextNode())) {
    const text = node as Text
    if (!text.data.trim() || text.parentElement?.closest('code, pre, script, style, textarea, [contenteditable="true"], [data-no-translate], .material-symbols-outlined, .font-code, .font-mono')) continue
    const previous = nodeTranslations.get(text)
    const source = previous && text.data === previous.rendered ? previous.english : text.data
    const translated = translateValue(source, locale)
    if (translated === null) {
      if (previous && text.data === previous.rendered && previous.rendered !== previous.english) text.data = previous.english
      nodeTranslations.delete(text)
      continue
    }
    const leading = source.match(/^\s*/)?.[0] || ''
    const trailing = source.match(/\s*$/)?.[0] || ''
    const rendered = `${leading}${translated}${trailing}`
    if (text.data !== rendered) text.data = rendered
    nodeTranslations.set(text, { english: source, rendered })
  }

  for (const element of [root, ...Array.from(root.querySelectorAll<HTMLElement>('*'))]) {
    if (element.matches('[contenteditable="true"], [data-no-translate]')) continue
    let saved = attributeTranslations.get(element)
    for (const attribute of translatedAttributes) {
      if (!element.hasAttribute(attribute)) continue
      const value = element.getAttribute(attribute) || ''
      const previous = saved?.get(attribute)
      const source = previous && value === previous.rendered ? previous.english : value
      const translated = translateValue(source, locale)
      if (translated === null) continue
      if (!saved) {
        saved = new Map()
        attributeTranslations.set(element, saved)
      }
      if (value !== translated) element.setAttribute(attribute, translated)
      saved.set(attribute, { english: source, rendered: translated })
    }
  }
}

export function localizePage(node: HTMLElement, locale: Locale) {
  let currentLocale = locale
  localizeTree(node, currentLocale)
  const observer = new MutationObserver((records) => {
    for (const record of records) {
      if (record.type === 'characterData' && record.target.parentElement?.isConnected) {
        const parent = record.target.parentElement
        if (parent.closest('code, pre, script, style, textarea, [contenteditable="true"], [data-no-translate], .material-symbols-outlined, .font-code, .font-mono')) continue
        const text = record.target as Text
        const previous = nodeTranslations.get(text)
        const source = previous && text.data === previous.rendered ? previous.english : text.data
        const translated = translateValue(source, currentLocale)
        if (translated !== null) {
          const leading = source.match(/^\s*/)?.[0] || ''
          const trailing = source.match(/\s*$/)?.[0] || ''
          const rendered = `${leading}${translated}${trailing}`
          if (text.data !== rendered) text.data = rendered
          nodeTranslations.set(text, { english: source, rendered })
        }
      } else if (record.type === 'childList') {
        for (const added of record.addedNodes) {
          if (added.nodeType === Node.ELEMENT_NODE) localizeTree(added as HTMLElement, currentLocale)
          else if (added.nodeType === Node.TEXT_NODE) {
            const text = added as Text
            if (text.parentElement?.closest('code, pre, script, style, textarea, [contenteditable="true"], [data-no-translate], .material-symbols-outlined, .font-code, .font-mono')) continue
            const source = text.data
            const translated = translateValue(source, currentLocale)
            if (translated !== null) {
              const leading = source.match(/^\s*/)?.[0] || ''
              const trailing = source.match(/\s*$/)?.[0] || ''
              const rendered = `${leading}${translated}${trailing}`
              nodeTranslations.set(text, { english: source, rendered })
              text.data = rendered
            }
          }
        }
      } else if (record.type === 'attributes') {
        localizeTree(record.target as HTMLElement, currentLocale)
      }
    }
  })
  observer.observe(node, { childList: true, characterData: true, subtree: true, attributes: true, attributeFilter: translatedAttributes })
  return {
    update(locale: Locale) {
      currentLocale = locale
      localizeTree(node, currentLocale)
    },
    destroy() {
      observer.disconnect()
    },
  }
}
