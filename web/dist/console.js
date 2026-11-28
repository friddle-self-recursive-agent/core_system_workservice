var CoreConsoleWeb = (() => {
  var __defProp = Object.defineProperty;
  var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
  var __getOwnPropNames = Object.getOwnPropertyNames;
  var __hasOwnProp = Object.prototype.hasOwnProperty;
  var __defNormalProp = (obj, key, value) => key in obj ? __defProp(obj, key, { enumerable: true, configurable: true, writable: true, value }) : obj[key] = value;
  var __export = (target, all) => {
    for (var name in all)
      __defProp(target, name, { get: all[name], enumerable: true });
  };
  var __copyProps = (to, from, except, desc) => {
    if (from && typeof from === "object" || typeof from === "function") {
      for (let key of __getOwnPropNames(from))
        if (!__hasOwnProp.call(to, key) && key !== except)
          __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
    }
    return to;
  };
  var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);
  var __publicField = (obj, key, value) => __defNormalProp(obj, typeof key !== "symbol" ? key + "" : key, value);

  // src/main.js
  var main_exports = {};
  __export(main_exports, {
    mount: () => mount
  });

  // src/core/i18n-en.js
  var enEntries = [
    /* ── Batch 3 IA（2026-11-26 全页扫描）：4 主线 NAV + 系统底层 + L-07 动态词 ── */
    ["Core Console", "Core Console"],
    ["\u9996\u9875", "Home"],
    ["\u5DE5\u4F5C\u4E0E\u6D41\u7A0B", "Work & Flows"],
    ["\u6D41\u7A0B\u8FD0\u884C", "Workflow Runs"],
    ["\u5BA1\u6279\u4E2D\u5FC3", "Approvals"],
    ["\u6587\u6863\u4E0E\u8D44\u4EA7", "Docs & Assets"],
    ["\u7CFB\u7EDF\u5E95\u5C42", "System Internals"],
    ["\u7CFB\u7EDF\u5DE5\u4F5C\u53F0", "System Workspace"],
    ["\u4E8B\u4EF6\u6D41", "Events"],
    ["\u63D0\u6848", "Proposals"],
    ["\u5173\u6CE8\u62A5\u544A", "Attention Report"],
    ["\u6279\u51C6\u7EE7\u7EED", "Approve & continue"],
    ["\u9A73\u56DE\u7EC8\u6B62", "Reject & stop"],
    ["\u8FD0\u884C\u8BE6\u60C5", "Run detail"],
    ["\u5F85\u6279", "Pending"],
    ["\u8FDB\u884C\u4E2D", "In progress"],
    ["\u6B63\u5728\u505A", "Doing"],
    ["\u7B49\u4EBA", "Waiting on you"],
    ["\u5361\u4F4F", "Stuck"],
    ["\u6392\u961F", "Queued"],
    ["\u9A8C\u8BC1\u4E2D", "Verifying"],
    ["\u7B49\u5F85\u62C9\u8D77", "Starting"],
    ["\u6392\u961F\u4E2D", "Queued"],
    ["\u5DF2\u5B8C\u6210", "Done"],
    ["\u5DF2\u53D6\u6D88", "Canceled"],
    ["\u5931\u8D25", "Failed"],
    ["\u53D7\u963B", "Blocked"],
    ["\u540C\u6B65\u4E2D", "Syncing"],
    ["\u5DF2\u540C\u6B65", "Synced"],
    ["\u6682\u65E0\u6D41\u7A0B\u6A21\u677F", "No workflow templates yet"],
    ["2s \u524D", "2s ago"],
    /* 侧栏/壳（main.js） */
    ["Core Console", "Core Console"],
    ["web \u91CD\u6784\u8D77\u70B9", "web rebuild baseline"],
    ["\u8FDE\u63A5\u4E2D\u2026", "Connecting\u2026"],
    ["\u5B9E\u65F6\uFF08WS patch\uFF09", "Live (WS patch)"],
    ["\u91CD\u8FDE\u4E2D\u2026", "Reconnecting\u2026"],
    ["\u964D\u7EA7\u8F6E\u8BE2\uFF0830s\uFF09", "Degraded polling (30s)"],
    ["\u4E2D\u6587", "\u4E2D\u6587"],
    ["EN", "EN"],
    /* welcome */
    ["\u5FEB\u7167\u88C5\u8F7D\u4E2D\u2026", "Loading snapshot\u2026"],
    ["\u6B63\u5728\u5EFA\u7ACB /v1/snapshot \u57FA\u7EBF\uFF08\u8FDE\u63A5\u5EFA\u7ACB\u5373\u63A8\u5168\u91CF\uFF09", "Establishing /v1/snapshot baseline (full push on connect)"],
    ["\u5728\u518C\u5BF9\u8C61", "Registered objects"],
    ["\u5168\u90E8 kind \u5408\u8BA1", "All kinds total"],
    ["kinds", "kinds"],
    ["scheme \u6CE8\u518C", "scheme registry"],
    ["\u4E8B\u4EF6\uFF08\u7A97\u53E3\u5185\uFF09", "Events (window)"],
    ["\u5F52\u6863\u884C", "archived rows"],
    ["\u5F85\u5BA1", "Pending review"],
    ["\u5F85\u5BA1\u533A\u4E3A\u7A7A", "No pending reviews (explicit empty state)"],
    ["\u6536\u4EF6\u7BB1", "Inbox"],
    ["\u5F85\u4F60\u5904\u7406", "awaiting you"],
    ["\u8FDB\u884C\u4E2D\u4EFB\u52A1", "Running tasks"],
    ["\u4EFB\u52A1\u4E2D\u5FC3", "Task Center"],
    ["kinds \u6458\u8981", "kinds summary"],
    ["\u9996\u4E2A\uFF1A", "first: "],
    [" \u4E2A", " item(s)"],
    /* events */
    ["\u4E8B\u4EF6\u6D41", "Event stream"],
    ["\u5F52\u6863\u4E8B\u4EF6\uFF08\u5FEB\u7167\u805A\u5408 + patch \u589E\u91CF\uFF1B\u70B9\u51FB\u770B\u5168\u6587\u4E0E\u94FE\u8DEF\uFF09", "Archived events (snapshot aggregate + patch increments; click for detail & chain)"],
    ["\u5168\u90E8\u7EA7\u522B", "All levels"],
    ["\u5FAA\u73AF\u5524\u9192", "Loop wake"],
    ["\u4E8B\u4EF6\u8BE6\u60C5", "Event detail"],
    ["\u539F\u59CB\u8F7D\u8377\uFF08JSON\uFF09", "Raw payload (JSON)"],
    ["\u2190 \u56DE\u4E8B\u4EF6\u6D41", "\u2190 Back to events"],
    ["\u5173\u8054", "Linked"],
    ["\u65E0", "none"],
    /* tasks */
    ["\u5206\u7EC4=\u786E\u5B9A\u6027\u6295\u5F71\uFF08\u72B6\u6001\u5355\u4E00\u5199\u8005=\u8C03\u5EA6\u5668\uFF09\uFF1B\u70B9\u51FB\u884C\u8FDB\u8BE6\u60C5", "Groups = deterministic projection (single status writer = scheduler); click a row for detail"],
    ["\u6B63\u5728\u505A", "Running"],
    ["\u7B49\u4EBA", "Waiting on human"],
    ["\u5361\u4F4F", "Stuck"],
    ["\u5B8C\u6210", "Done"],
    ["\u6392\u961F", "Queued"],
    ["\u9879", "item(s)"],
    ["\u6B63\u5728\u505A\uFF1A\u6682\u65E0\u3002", "Running: none (explicit empty state)."],
    ["\u7B49\u4EBA\uFF1A\u6682\u65E0\u3002", "Waiting on human: none (explicit empty state)."],
    ["\u5361\u4F4F\uFF1A\u6682\u65E0\u3002", "Stuck: none (explicit empty state)."],
    ["\u5B8C\u6210\uFF1A\u6682\u65E0\u3002", "Done: none (explicit empty state)."],
    ["\u6392\u961F\uFF1A\u6682\u65E0\u3002", "Queued: none (explicit empty state)."],
    ["\u72B6\u6001\u672A\u77E5", "state unknown"],
    ["\u6700\u8FD1\u52A8\u4F5C\uFF1A", "last action: "],
    ["\u5173\u6CE8\u62A5\u544A", "Attention Report"],
    ["\u4EFB\u52A1\u8BE6\u60C5", "Task detail"],
    ["\u2190 \u56DE\u4EFB\u52A1\u4E2D\u5FC3", "\u2190 Back to tasks"],
    ["\u8BC1\u636E\u94FE", "Evidence chain"],
    /* objects */
    ["\u5BF9\u8C61\u5DE5\u4F5C\u53F0", "Object Workbench"],
    ["\u8BE5 kind \u6682\u65E0\u5BF9\u8C61", "No objects for this kind"],
    ["\u6682\u65E0\u5728\u518C\u5BF9\u8C61", "No registered objects"],
    ["\u5FAA\u73AF", "Loop"],
    ["\u76F8\u4F4D", "Phase"],
    ["\u8D1F\u8D23", "Owner"],
    ["\u5F00\u59CB\u4E8E", "started at"],
    ["\u76F8\u4F4D\u673A", "Phase machine"],
    ["\u4E8B\u4EF6\u94FE", "Event chain"],
    ["\u5BF9\u8C61\u8BE6\u60C5", "Object detail"],
    ["\u4EE3\u7801\u5BA1\u6838", "Code Audit"],
    ["\u6807\u7B7E", "Tags"],
    /* 「补标/Add tag」幽灵词条已销（V-22，2026-11-26 概念可见性批）：词表对账零 UI 引用；
     * tag 声明通路=对象 spec.metadata.tags 经 Apply 落库（聚合查看=新「标签」屏——非表单录入面）。 */
    ["\u672A\u5206\u7EC4", "Ungrouped"],
    ["\u4E0A\u4E0B\u6587", "Context"],
    ["\u5DE5\u4F5C\u53F0", "Workbench"],
    ["\u5386\u53F2", "History"],
    ["\u6B63\u6587", "Content"],
    ["\u7F16\u8F91", "Edit"],
    ["\u4FDD\u5B58", "Save"],
    ["\u53D6\u6D88", "Cancel"],
    ["\u8349\u7A3F", "Draft"],
    ["\u5E94\u7528", "Apply"],
    ["\u5F03\u7A3F", "Discard draft"],
    ["\u4E50\u89C2\u9501", "optimistic lock"],
    /* chat / sessionctx */
    ["\u4F1A\u8BDD", "Sessions"],
    ["\u4F1A\u8BDD\u6E05\u5355", "Session list"],
    ["\u4E0E root \u5BF9\u8BDD", "Chat with root"],
    ["\u5BF9\u8BDD", "Chat"],
    ["\u8F6C\u5F55", "Transcript"],
    ["\u8F6C\u5F55\u9762\u672A\u88C5\u914D\u2014\u2014\u5BBF\u4E3B\u6CE8\u5165 window.__CC.chatFaces \u540E\u73B0\uFF1B\u4F1A\u8BDD\u539F\u6587\u4E0D\u5165\u516C\u5171\u5E93\uFF08intake \u6682\u5047\u9762\u4E3A\u51C6\uFF09\u3002", "Transcript face not mounted \u2014 appears after host injects window.__CC.chatFaces; raw session text stays out of the public library (intake placeholder for now)."],
    ["\u5165\u53E3", "Entry points"],
    ["\u76F8\u5173\u4F1A\u8BDD", "Related sessions"],
    ["\u4E0A\u4E0B\u6587\uFF08ContextPack \u5951\u7EA6\u9762\uFF09", "Context (ContextPack contract face)"],
    ["\u7D22\u5F15\u5019\u9009\u6295\u5F71", "index-candidate projection"],
    ["\u5199\u9762\u5F00", "Writes open"],
    ["\u5199\u9762\u5173\u65AD", "Writes closed"],
    ["\u5FC3\u8DF3", "Heartbeats"],
    ["\u6B21 \xB7 \u6700\u8FD1", "hit(s) \xB7 last"],
    ["\u52A8\u4F5C", "Action"],
    ["\u6CE8\u5165\u56FA\u5B9A prompt\u300C\u7EE7\u7EED\u300D\uFF08\u81EA\u6108\u7B2C\u4E00\u52A8\u4F5C\uFF09", "Inject fixed prompt \u300C\u7EE7\u7EED\u300D (self-heal first action)"],
    ["\u7F3A\u7701", "default"],
    ["\u57DF", "range"],
    ["\u672A\u88C5\u914D\uFF08tagSigSource=none\uFF09", "not assembled (tagSigSource=none)"],
    ["\u5206\u7247\u7F13\u5B58", "Shard cache"],
    ["\u5171", "total "],
    ["\u7247 \xB7 stale", " shard(s) \xB7 stale"],
    ["\u91CD\u5EFA\u9608\u503C", "rebuild threshold"],
    ["\u65E0\u8986\u76D6\uFF08\u751F\u6548\u503C=ContextPolicy \u7F3A\u7701\u2014\u2014PUT \u540E\u6B64\u5904\u73B0\u8BFB\u8986\u76D6\u884C\uFF09\u3002", "No override (effective value = ContextPolicy default \u2014 PUT and the override row appears here)."],
    ["\u4FDD\u5B58\u5206\u53C9", "Save fork"],
    ["\u64A4\u9500\u4E0A\u6B21", "Undo last"],
    ["\u5206\u53C9\u5907\u6CE8\uFF08\u53EF\u9009\uFF09", "Fork note (optional)"],
    ["tag \u96C6\uFF1AtagType:tag \u9017\u53F7/\u7A7A\u683C\u5206\u9694\uFF08\u6BCF\u7C7B\u22645 \u603B\u226420\uFF09", "tag set: tagType:tag, comma/space separated (\u22645 per type, \u226420 total)"],
    ["chunks\uFF08\u7D22\u5F15 TopK \u5019\u9009\uFF09", "chunks (index TopK candidates)"],
    ["\u65E0\u7D22\u5F15\u5019\u9009\uFF08tagSig \u672A\u88C5\u914D\u6216\u7D22\u5F15\u96F6\u547D\u4E2D\u2014\u2014\u4E0D\u5192\u5145\u88C5\u914D\u5305\u5185\u5BB9\uFF09\u3002", "No index candidates (tagSig not assembled or zero hits \u2014 never fabricates pack content)."],
    ["\u76F8\u5173 Tag \u5206\u7247\uFF08TagCtxShard\uFF09", "Related tag shards (TagCtxShard)"],
    ["\u65E0\u751F\u6548 tag \u5206\u7247\uFF08tagSig \u672A\u88C5\u914D\u2014\u2014\u7F16\u8F91\u5206\u53C9\u540E\u9010 tagKey \u6295\u5F71 TagCtxShard\uFF09\u3002", "No active tag shards (tagSig not assembled \u2014 per-tagKey TagCtxShard projection after fork edit)."],
    ["\u8986\u76D6\u884C", "Override"],
    ["\u672C\u4EE3\u65E0\u4E09\u6BB5\u5F0F Session \u88C5\u914D\u5668\u2014\u2014chunks=\u7D22\u5F15 TopK \u5019\u9009\u6295\u5F71\uFF08context.md \u7D22\u5F15\u7B2C\u4E00\u6BB5\u8BCD\u5F62\uFF09\uFF0C\u975E\u4F1A\u8BDD\u771F\u5B9E\u88C5\u914D\u5305\u3002", "This generation has no three-stage Session assembler \u2014 chunks are index TopK candidate projections (context.md stage-one word form), not the real session pack."],
    /* approvals */
    ["\u5BA1\u6279", "Approvals"],
    ["\u4E24\u6B65\u6279\u51C6", "Two-step approval"],
    ["\u5F85\u5BA1\u533A", "Pending"],
    ["\u5DF2\u51B3\u8D26", "Decision ledger"],
    ["\u6279\u51C6", "Approve"],
    ["\u9A73\u56DE", "Reject"],
    ["\u6682\u65E0\u5F85\u5BA1\u9879\u3002", "No pending items (explicit empty state)."],
    ["\u6682\u65E0\u5DF2\u51B3\u8BB0\u5F55\u3002", "No decisions recorded (explicit empty state)."],
    /* workflows */
    ["\u5DE5\u4F5C\u6D41", "Workflows"],
    ["\u8FD0\u884C", "Run"],
    ["\u8FD0\u884C\u62BD\u5C49", "Run drawer"],
    ["\u89E6\u53D1", "Trigger"],
    ["\u6B65\u9AA4", "Steps"],
    ["\u6682\u65E0\u5DE5\u4F5C\u6D41\u3002", "No workflows (explicit empty state)."],
    /* misc */
    ["\u7CFB\u7EDF\u6811", "System Tree"],
    ["\u6587\u6863\u5E93", "Docs"],
    ["\u63D0\u6848", "Proposals"],
    ["\u8FDE\u63A5\u5C42", "Connectors"],
    ["\u7CFB\u7EDF", "System"],
    ["\u8FDE\u63A5\u5668", "Connectors"],
    ["\u8BBE\u8BA1\u6587\u6863", "Design docs"],
    ["\u6682\u65E0\u5185\u5BB9\u3002", "No content (explicit empty state)."],
    ["\u52A0\u8F7D\u5931\u8D25", "Load failed"],
    ["\u91CD\u8BD5", "Retry"],
    [" vs \u7F16\u8F91\u57FA\u7EBF r", " vs draft base r"],
    [" \xB7 \u57DF ", " \xB7 range "],
    [" \xB7 \u5907\u6CE8 ", " \xB7 note "],
    [" \xB7 \u622A\u81F3 ", " \xB7 as of "],
    [" \xB7 \u62A5\u544A\u4E3A\u6D3E\u751F\u89C6\u56FE\u53EF\u968F\u65F6\u91CD\u5EFA\uFF0C\u4E0D\u843D\u5BF9\u8C61\u8D26\u3002</div>", " \xB7 This report is a derived view and can be rebuilt anytime; not persisted as objects.</div>"],
    [" \xB7 \u6700\u8FD1 ", " \xB7 last "],
    [" \xB7 \u68C0\u67E5\u8303\u56F4 ", " \xB7 check scope "],
    [" \xB7 \u7A3F ", " \xB7 draft "],
    [" \xB7 \u7F3A\u5E2D\u6570\u636E\uFF1A", " \xB7 missing data: "],
    [" \xB7 \u91CD\u5EFA\u9608\u503C ", " \xB7 rebuild threshold "],
    [" \u2192 \u5DE5\u4F5C\u6D41", " \u2192 Workflow"],
    [" \u4E2A kind\u2014\u2014\u89C1\u5BF9\u8C61\u5DE5\u4F5C\u53F0\u5DE6\u4FA7\u6E05\u5355\u3002</div>", " kinds \u2014 see the list on the left of the Object Workbench.</div>"],
    [" \u4E2A\u76F8\u5173", " related"],
    [" \u4E2A\uFF1A", " item(s): "],
    [" \u4EFD\u89C4\u5219\u6587\u6863\uFF09\u3002</div>", " rule document(s))."],
    [" \u5904\u53D8\u66F4\uFF09", " change(s))"],
    [" \u603B\u2264", " total \u2264"],
    [" \u6B21", " time(s)"],
    [" \u6B21\u8FD0\u884C", " run(s)"],
    [" \u7247 \xB7 stale ", " shard(s) \xB7 stale "],
    [" \u7684\u5BF9\u8C61\u4E0E\u4E8B\u4EF6\uFF08\u5FEB\u7167\u805A\u5408\u6295\u5F71\uFF09", " objects & events (snapshot aggregate projection)"],
    [" \u7684\u6570\u503C\uFF08ContextPolicy \u9884\u7B97\u57DF\uFF09", " value (ContextPolicy budget range)"],
    [" \u7F3A", " missing"],
    [" \u884C\u2014\u2014\u7528\u8FC7\u6EE4\u6536\u7A84\u3002</div>", " rows \u2014 narrow with the filter.</div>"],
    [" \u884C\u5DEE\u5F02</div>", " line diff</div>"],
    [" \u9879", " item(s)"],
    [" \u9879</span>", " item(s)</span>"],
    ['" data-docs-kind="">\u5168\u90E8</span>', '" data-docs-kind="">All</span>'],
    ['" data-ed-tab="draft">\u8349\u7A3F</div>', '" data-ed-tab="draft">Draft</div>'],
    ['" data-ed-tab="history">\u5386\u53F2</div>', '" data-ed-tab="history">History</div>'],
    ['" data-ed-view-btn="raw">\u539F\u6587</span>', '" data-ed-view-btn="raw">Source</span>'],
    ['" data-ed-view-btn="render">\u6B63\u6587</span>', '" data-ed-view-btn="render">Rendered</span>'],
    ['" title="\u63D0\u53D6\u9884\u7B97\uFF08token \u6570\uFF09">', '" title="Take budget (token count)">'],
    ['" title="\u770B\u540C tag \u5BF9\u8C61">#', '" title="Objects with the same tag">#'],
    ['"><div class="small faint">diff \u73B0\u8BFB\u4E2D\u2026</div></div>', '"><div class="small faint">Loading diff\u2026</div></div>'],
    ['">\u5FFD\u7565</button> ', '">Dismiss</button> '],
    ['">\u5FFD\u7565</button></div></div></div>', '">Dismiss</button></div></div></div>'],
    ['">\u6240\u5C5E scope</button> ', '">Owning scope</button> '],
    ['">\u6279\u51C6\uFF08= Apply\uFF09</button> ', '">Approve (= Apply)</button> '],
    ['">\u7F16\u8F91\u5668</button>', '">Open editor</button>'],
    ['">\u7F16\u8F91\u5668</button> ', '">Open editor</button> '],
    ['</b> <span class="small faint">\uFF08\u7F3A\u7701 ', '</b> <span class="small faint">(default '],
    ['<button class="btn btn-secondary sm" data-ctx-save>\u4FDD\u5B58\u5206\u53C9</button>', '<button class="btn btn-secondary sm" data-ctx-save>Save fork</button>'],
    ['<details class="md-raw"><summary>\u539F\u59CB\u8F7D\u8377\uFF08JSON\uFF09</summary><pre class="mono md-raw-pre">', '<details class="md-raw"><summary>Raw payload (JSON)</summary><pre class="mono md-raw-pre">'],
    ['<div class="apply-dock"><button class="btn btn-primary block" data-ed-apply disabled>Apply\uFF08\u65E0\u53D8\u66F4\uFF09</button></div>', '<div class="apply-dock"><button class="btn btn-primary block" data-ed-apply disabled>Apply (no changes)</button></div>'],
    ['<div class="card"><div class="card-b"><div class="small faint">\u4E0A\u4E0B\u6587\u9762\u6682\u4E0D\u53EF\u8FBE\uFF08', '<div class="card"><div class="card-b"><div class="small faint">Context face unreachable ('],
    ['<div class="card"><div class="card-b"><div class="small faint">\u4F1A\u8BDD\u6E05\u5355\u6682\u4E0D\u53EF\u5F97', '<div class="card"><div class="card-b"><div class="small faint">Session list unavailable'],
    ['<div class="card"><div class="card-b"><div class="small faint">\u6536\u4EF6\u7BB1\u4E3A\u7A7A\uFF08\u96F6\u5F85\u529E\uFF09\u3002</div></div></div>', '<div class="card"><div class="card-b"><div class="small faint">Inbox is empty (zero to-dos \u2014 explicit empty state).</div></div></div>'],
    ['<div class="card"><div class="card-b"><div class="small faint">\u6536\u4EF6\u7BB1\u6682\u4E0D\u53EF\u7528\uFF08\u6295\u5F71\u7F3A\u5E2D\uFF09\u3002</div></div></div>', '<div class="card"><div class="card-b"><div class="small faint">Inbox projection absent (snapshot carries no /v1/inbox aggregate) \u2014 your to-do face is explicitly absent, never faked.</div></div></div>'],
    ['<div class="card"><div class="card-h"><span class="h-title">\u6982\u89C8</span></div><div class="card-b">', '<div class="card"><div class="card-h"><span class="h-title">Overview</span></div><div class="card-b">'],
    ['<div class="card-h"><span class="h-title">\u4E0A\u4E0B\u6587\uFF08ContextPack \u5951\u7EA6\u9762\uFF09</span><span class="h-right">', '<div class="card-h"><span class="h-title">Context (ContextPack contract face)</span><span class="h-right">'],
    ['<div class="card-h"><span class="h-title">\u5386\u53F2</span></div><div class="card-b">', '<div class="card-h"><span class="h-title">History</span></div><div class="card-b">'],
    ['<div class="card-h"><span class="h-title">\u8349\u7A3F</span></div><div class="card-b">', '<div class="card-h"><span class="h-title">Draft</span></div><div class="card-b">'],
    ['<div class="card-h"><span class="h-title">\u8349\u7A3F</span><span class="h-right">', '<div class="card-h"><span class="h-title">Draft</span><span class="h-right">'],
    ['<div class="chat-grid"><div><h2 class="sec-title">\u4F1A\u8BDD\u6E05\u5355</h2>', '<div class="chat-grid"><div><h2 class="sec-title">Session list</h2>'],
    ['<div class="dd-grid"><div data-dd-tree><div class="small faint">\u4ED3\u6811\u73B0\u8BFB\u4E2D\u2026</div></div>', '<div class="dd-grid"><div data-dd-tree><div class="small faint">Loading repo tree\u2026</div></div>'],
    ['<div class="diff-line ctx">\u2026 \u5171 ', '<div class="diff-line ctx">\u2026 total '],
    ['<div class="mr-sub small">\u76F8\u5173\u5BF9\u8C61 ', '<div class="mr-sub small">Related objects '],
    ['<div class="ob-empty">scheme \u6682\u4E0D\u53EF\u8FBE\uFF08kinds \u805A\u5408\u7F3A\u5E2D\uFF09\u3002</div>', '<div class="ob-empty">scheme unreachable (kinds aggregate absent).</div>'],
    ['<div class="ob-empty">\u4E8B\u4EF6\u5168\u6587\u6682\u4E0D\u53EF\u8FBE\uFF08', '<div class="ob-empty">Event detail unreachable ('],
    ['<div class="ob-empty">\u4ED3\u6811\u4E3A\u7A7A\u3002</div>', '<div class="ob-empty">Repo tree is empty (honest empty state).</div>'],
    ['<div class="ob-empty">\u4ED3\u6811\u6682\u4E0D\u53EF\u8BFB\uFF08', '<div class="ob-empty">Repo tree unreadable ('],
    ['<div class="ob-empty">\u65E0\u9002\u7528\u89C4\u5219\u2014\u2014\u6309 kind/scope/tag/\u6B63\u6587\u5F15\u7528\u56DB\u8DEF\u68C0\u67E5\u96F6\u547D\u4E2D\uFF08\u68C0\u67E5\u8303\u56F4=', '<div class="ob-empty">No applicable rules \u2014 zero hits across kind/scope/tag/content-reference checks (check scope ='],
    ['<div class="ob-empty">\u6682\u65E0\u6807\u7B7E\u2014\u2014\u5BF9\u8C61\u6B63\u6587\u672A\u58F0\u660E spec.metadata.tags\uFF08\u58F0\u660E\u7ECF Apply \u843D\u5E93\u540E\u6B64\u5904\u73B0\u8BFB\uFF09\u3002</div>', '<div class="ob-empty">No tags \u2014 object content declares no spec.metadata.tags (appears here once applied via Apply).</div>'],
    ['<div class="ob-empty">\u672C\u6587\u6863\u6682\u65E0\u6B63\u6587\u3002</div>', '<div class="ob-empty">This document has no content.</div>'],
    ['<div class="ob-empty">\u6B63\u6587\u6682\u4E0D\u53EF\u8BFB\uFF08', '<div class="ob-empty">Content unreadable ('],
    ['<div class="ob-empty">\u8BE5 kind \u65E0\u5B57\u6BB5\u58F0\u660E\uFF08spec \u7F3A\u5E2D\uFF09\u3002</div>', '<div class="ob-empty">This kind has no field declarations (no spec).</div>'],
    ['<div class="ob-empty">\u8F6C\u5F55\u9762\u672A\u88C5\u914D\u2014\u2014\u5BBF\u4E3B\u6CE8\u5165 window.__CC.chatFaces \u540E\u73B0\u8BFB\uFF1B', '<div class="ob-empty">Transcript face not mounted \u2014 reads live once the host injects window.__CC.chatFaces;'],
    ['<div class="ob-empty">\uFF08\u7A7A\u58F0\u660E\uFF09</div>', '<div class="ob-empty">(empty declaration)</div>'],
    ['<div class="report-foot small faint">\u62A5\u544A\u8BF4\u660E\uFF1A', '<div class="report-foot small faint">Data confidence:'],
    ['<div class="small faint" style="grid-column:1/-1">\u2026\u53E6\u6709 ', '<div class="small faint" style="grid-column:1/-1">\u2026and '],
    ['<div class="small faint" style="margin-bottom:10px">\u4EE5\u4E0B\u4E8B\u9879\u6765\u81EA\u6700\u8FD1\u4E8B\u4EF6\u3002</div>', '<div class="small faint" style="margin-bottom:10px">Snapshot carries no attention aggregate \u2014 this card is projected from archived event rows (ApprovalNeeded/Escalate).</div>'],
    ['<div class="small faint" style="margin-bottom:10px">\u5F53\u524D\u6CA1\u6709\u9700\u8981\u4F60\u5173\u6CE8\u7684\u4E8B\u9879\u3002</div>', '<div class="small faint" style="margin-bottom:10px">No attention aggregate in snapshot and no approval/escalation rows in the archive \u2014 this report converges to zero hits on recorded facts (never faked).</div>'],
    ['<div class="small faint" style="margin-bottom:6px">\u7A3F\u6765\u6E90 fieldManager\uFF1A', '<div class="small faint" style="margin-bottom:6px">Draft source fieldManager: '],
    ['<div class="small faint" style="margin-top:6px">\u57FA\u51C6=r', '<div class="small faint" style="margin-top:6px">base=r'],
    ['<div class="small faint" style="margin-top:6px">\u65E0\u8986\u76D6\uFF08\u751F\u6548\u503C=ContextPolicy \u7F3A\u7701\u2014\u2014PUT \u540E\u6B64\u5904\u73B0\u8BFB\u8986\u76D6\u884C\uFF09\u3002</div>', '<div class="small faint" style="margin-top:6px">No override (effective value = ContextPolicy default \u2014 PUT and the override row appears here).</div>'],
    ['<div class="small faint" style="margin-top:6px">\u73B0\u884C\u8986\u76D6\uFF1Arevision ', '<div class="small faint" style="margin-top:6px">Current override: revision '],
    ['<div class="small faint" style="margin-top:6px">\u7A3F\u6765\u6E90 fieldManager\uFF1A', '<div class="small faint" style="margin-top:8px">Draft source fieldManager: '],
    ['<div class="small faint" style="margin-top:8px">\u8F6C\u5F55\u4E0E\u7EC8\u7AEF\u9762\u7531\u5BBF\u4E3B faces \u6CE8\u5165\uFF08__CC.chatFaces / __CC.termFaces\uFF09\uFF1B', '<div class="small faint" style="margin-top:8px">Transcript & terminal faces are host-injected (__CC.chatFaces / __CC.termFaces); '],
    ['<div class="small faint" style="padding:8px 14px">\u2026\u5171 ', '<div class="small faint" style="padding:8px 14px">\u2026total '],
    ['<div class="small faint">\u4F1A\u8BDD\u4E0D\u5728\u6D3B\u6027\u8D26\uFF08\u65E0\u5FC3\u8DF3\u884C\uFF09\u2014\u2014\u5951\u7EA6\u9762\u4ECD\u53EF\u67E5\u770B/\u7F16\u8F91\u672C\u4F1A\u8BDD\u4E0A\u4E0B\u6587\u8986\u76D6\uFF08\u4E0D\u5192\u5145\u5728\u518C\uFF09\u3002</div>', `<div class="small faint">Session not in the liveness ledger (no heartbeat row) \u2014 the contract face still shows/edits this session's context override (never pretends it is live).</div>`],
    ['<div class="small faint">\u4FDD\u5B58\u5931\u8D25\uFF08', '<div class="small faint">Save failed ('],
    ['<div class="small faint">\u5199\u9762\u5173\u65AD\uFF08--allow-apply \u672A\u5F00\u542F\uFF09\u2014\u2014\u7F16\u8F91\u5206\u53C9\u4E0D\u53EF\u7528\uFF08\u4E0E\u5FC3\u8DF3\u6CE8\u518C\u9762\u540C\u5F00\u5173\uFF0C\u4E0D\u5192\u5145\u53EF\u7F16\u8F91\uFF09\u3002</div>', '<div class="small faint">Writes closed (--allow-apply off) \u2014 fork editing unavailable (same switch as the heartbeat registration face; never pretends to be editable).</div>'],
    ['<div class="small faint">\u5206\u7247\u7F13\u5B58\u9762\u7F3A\u5E2D\uFF08contextidx \u672A\u66B4\u9732 Shards\u2014\u2014\u8BDA\u5B9E\u964D\u7EA7\uFF0C\u4E0D\u5192\u5145\u7A7A\u7F13\u5B58\uFF09\u3002</div>', '<div class="small faint">Shard cache face absent (contextidx exposes no Shards \u2014 honest degradation, never pretends an empty cache).</div>'],
    ['<div class="small faint">\u5951\u7EA6\u672A\u643A\u5E26\u6B65\u9AA4\u660E\u7EC6\u2014\u2014\u4E0D\u663E\u793A\u4E0D\u7F16\u9020\u3002</div>', '<div class="small faint">Contract carries no step details \u2014 hidden, never fabricated.</div>'],
    ['<div class="small faint">\u5BA1\u6279\u8BB0\u5F55\u6682\u4E0D\u53EF\u8BFB\uFF08', '<div class="small faint">Approval records unreadable ('],
    ['<div class="small faint">\u5DF2\u4FDD\u5B58\uFF08revision ', '<div class="small faint">Saved (revision '],
    ['<div class="small faint">\u6682\u65E0\u5DF2\u51B3\u8BB0\u5F55\uFF08\u6279\u51C6/\u9A73\u56DE\u540E\u6B64\u5904\u663E\u793A\uFF09\u3002</div>', '<div class="small faint">No decision rows yet (approvals/rejections appear immediately).</div>'],
    ['<div class="small faint">\u5DF2\u64A4\u9500\u2014\u2014\u5199\u524D\u5FEB\u7167\u5DF2\u6062\u590D\u3002</div>', '<div class="small faint">Reverted \u2014 pre-write snapshot restored.</div>'],
    ['<div class="small faint">\u64A4\u9500\u5931\u8D25\uFF08', '<div class="small faint">Revert failed ('],
    ['<div class="small faint">\u65E0\u751F\u6548 tag \u5206\u7247\uFF08tagSig \u672A\u88C5\u914D\u2014\u2014\u7F16\u8F91\u5206\u53C9\u540E\u9010 tagKey \u6295\u5F71 TagCtxShard\uFF09\u3002</div>', '<div class="small faint">No active tag shards (tagSig not assembled \u2014 per-tagKey TagCtxShard projection after a fork edit).</div>'],
    ['<div class="small faint">\u65E0\u7D22\u5F15\u5019\u9009\uFF08tagSig \u672A\u88C5\u914D\u6216\u7D22\u5F15\u96F6\u547D\u4E2D\u2014\u2014\u4E0D\u5192\u5145\u88C5\u914D\u5305\u5185\u5BB9\uFF09\u3002</div>', '<div class="small faint">No index candidates (tagSig not assembled or zero index hits \u2014 never fabricates pack content).</div>'],
    ['<div class="small faint">\u65E0\u8349\u7A3F\u2014\u2014\u7F16\u8F91\u300C\u539F\u6587\u300D\u5373\u7D2F\u79EF\u8349\u7A3F\uFF08\u5199\u7A7F /v1/drafts\uFF0C\u4E0D\u52A8\u6B63\u5F0F\u6587\u6863\u3001\u4E0D\u53D1\u4E8B\u4EF6\uFF09\u3002</div>', '<div class="small faint">No draft \u2014 editing "Source" accumulates a draft (writes through /v1/drafts; never touches the official document, no events).</div>'],
    ['<div class="small faint">\u6B63\u6587\u6682\u4E0D\u53EF\u8BFB\uFF08', '<div class="small faint">Content unreadable ('],
    ['<div class="small faint">\u6B63\u6587\u73B0\u8BFB\u4E2D\u2026</div>', '<div class="small faint">Loading content\u2026</div>'],
    ['<div class="small faint">\u89C4\u5219\u547D\u4E2D\u5224\u5B9A\u4E2D\u2026</div>', '<div class="small faint">Evaluating rule hits\u2026</div>'],
    ['<div class="small faint">\u8BE5 kind \u6682\u65E0\u5728\u518C\u5BF9\u8C61\u3002</div>', '<div class="small faint">No registered objects for this kind (explicit empty state).</div>'],
    ['<div class="small faint">\u8BE5 workflow \u6682\u65E0\u8FD0\u884C\u8BB0\u5F55\u3002</div>', '<div class="small faint">No runs for this workflow (explicit empty state).</div>'],
    ['<div class="small faint">\u8BE5\u5BF9\u8C61\u6682\u65E0\u5F52\u6863\u4E8B\u4EF6\u884C\uFF08DocApplied/DocReview \u7B49\uFF09\u3002</div>', '<div class="small faint">No archived event rows for this object (DocApplied/DocReview etc. \u2014 explicit empty state).</div>'],
    ['<div class="small faint">\u9ED8\u8BA4\u5C55\u793A\u6E32\u67D3\u540E\u7684\u6210\u54C1\uFF1B\u70B9\u300C\u539F\u6587\u300D\u67E5\u770B\u4E0E\u7F16\u8F91\u6E90\u6587\u4EF6\u3002\u7F16\u8F91\u5373\u7D2F\u79EF\u8349\u7A3F\uFF08POST /v1/drafts\u2014\u2014\u4E0D\u52A8\u6B63\u5F0F\u6587\u6863\u3001\u4E0D\u53D1\u4E8B\u4EF6\uFF09\uFF1BApply \u552F\u4E00\u5165\u53E3=\u53F3\u680F\u5438\u5E95\u6309\u94AE\u3002</div>', '<div class="small faint">Shows the rendered result by default; click "Source" to view/edit the raw file. Editing accumulates a draft (POST /v1/drafts \u2014 never touches the official document, no events); the only Apply entry is the docked button in the right column.</div>'],
    ['<div class="small faint">YAML \u5BF9\u8C61\u9ED8\u8BA4\u5373\u539F\u6587\uFF08\u7EAF\u6587\u672C\u5168\u6587\uFF09\uFF0C\u65E0\u9700\u70B9\u300C\u539F\u6587\u300D\u67E5\u770B\uFF1B\u4FEE\u6539\u4F18\u5148\u8D70\u300C\u5BF9\u8BDD\u4FEE\u6539\u300D\u4E3B\u901A\u8DEF\uFF08LLM \u6539\u6587\u6863\u2192Apply\uFF09\uFF0C\u624B\u6539\u539F\u6587=\u6700\u540E\u624B\u6BB5\u3002\u7F16\u8F91\u5373\u7D2F\u79EF\u8349\u7A3F\uFF08POST /v1/drafts\u2014\u2014\u4E0D\u52A8\u6B63\u5F0F\u6587\u6863\u3001\u4E0D\u53D1\u4E8B\u4EF6\uFF09\uFF1BApply \u552F\u4E00\u5165\u53E3=\u53F3\u680F\u5438\u5E95\u6309\u94AE\u3002</div>', '<div class="small faint">YAML objects show raw source by default (plain-text full text) \u2014 no need to click "Source". Prefer chat-edit as the primary path (LLM edits the doc \u2192 Apply); hand-editing raw text is a last resort. Editing accumulates a draft (POST /v1/drafts \u2014 never touches the official document, no events); the only Apply entry is the docked button in the right column.</div>'],
    ['<div class="w-foot small faint">\u6570\u636E\u9762\uFF1AGET /v1/snapshot \u57FA\u7EBF + /v1/stream patch \u589E\u91CF\uFF08WS \u65AD\u7EBF\u81EA\u52A8\u964D\u7EA7\u8F6E\u8BE2\uFF09\uFF1B', '<div class="w-foot small faint">Data plane: GET /v1/snapshot baseline + /v1/stream patch increments (auto-degrades to polling on WS loss); '],
    ['<div data-ap-ledger style="margin-top:20px"><div class="small faint">\u5DF2\u51B3\u8BB0\u5F55\u73B0\u8BFB\u4E2D\u2026</div></div>', '<div data-ap-ledger style="margin-top:20px"><div class="small faint">Loading decision ledger\u2026</div></div>'],
    ['<div data-ctx-panel><div class="card"><div class="card-b"><div class="small faint">\u4E0A\u4E0B\u6587\u52A0\u8F7D\u4E2D\u2026\uFF08GET /v1/sessions/', '<div data-ctx-panel><div class="card"><div class="card-b"><div class="small faint">Hydrating context\u2026 (GET /v1/sessions/'],
    ['<div><h2 class="sec-title">\u8F6C\u5F55</h2><div class="card"><div class="card-b transcript">', '<div><h2 class="sec-title">Transcript</h2><div class="card"><div class="card-b transcript">'],
    ['<h2 class="sec-title" style="margin-top:14px">\u4E0A\u4E0B\u6587</h2>', '<h2 class="sec-title" style="margin-top:14px">Context</h2>'],
    ['<h2 class="sec-title">kinds \u6458\u8981</h2>', '<h2 class="sec-title">kinds summary</h2>'],
    ['<h2 class="sec-title">\u4F1A\u8BDD\u6E05\u5355</h2>', '<h2 class="sec-title">Session list</h2>'],
    ['<h2 class="sec-title">\u5165\u53E3</h2>', '<h2 class="sec-title">Entry points</h2>'],
    ['<h2 class="sec-title">\u5BF9\u8C61</h2><div class="card">', '<h2 class="sec-title">Object</h2><div class="card">'],
    ['<h2 class="sec-title">\u5DF2\u51B3\u8BB0\u5F55</h2><div style="overflow-x:auto"><table class="tbl"><thead><tr>', '<h2 class="sec-title">Decision ledger</h2><div style="overflow-x:auto"><table class="tbl"><thead><tr>'],
    ['<h2 class="sec-title">\u5F53\u524D\u5FAA\u73AF</h2>', '<h2 class="sec-title">Current loop</h2>'],
    ['<h2 class="sec-title">\u6536\u4EF6\u7BB1\uFF08\u5F85\u4F60\u5904\u7406\uFF09</h2>', '<h2 class="sec-title">Inbox (awaiting you)</h2>'],
    ['<h2 class="sec-title">\u6700\u8FD1\u76F8\u5173\u4E8B\u4EF6</h2><div class="card">', '<h2 class="sec-title">Recent related events</h2><div class="card">'],
    ['<h2 class="sec-title">\u6B65\u9AA4</h2><div class="card">', '<h2 class="sec-title">Steps</h2><div class="card">'],
    ['<h2 class="sec-title">\u8FD0\u884C\u8BB0\u5F55</h2>', '<h2 class="sec-title">Run history</h2>'],
    ['<h3 class="sec-title" style="margin:10px 0 6px">chunks\uFF08\u7D22\u5F15 TopK \u5019\u9009\uFF09</h3>', '<h3 class="sec-title" style="margin:10px 0 6px">chunks (index TopK candidates)</h3>'],
    ['<h3 class="sec-title" style="margin:10px 0 6px">\u76F8\u5173 Tag \u5206\u7247\uFF08TagCtxShard\uFF09</h3>', '<h3 class="sec-title" style="margin:10px 0 6px">Related tag shards (TagCtxShard)</h3>'],
    ['<input class="inp" data-ctx-note placeholder="\u5206\u53C9\u5907\u6CE8\uFF08\u53EF\u9009\uFF09">', '<input class="inp" data-ctx-note placeholder="Fork note (optional)">'],
    ['<input class="inp" data-ctx-tags placeholder="tag \u96C6\uFF1AtagType:tag \u9017\u53F7/\u7A7A\u683C\u5206\u9694\uFF08\u6BCF\u7C7B\u2264', '<input class="inp" data-ctx-tags placeholder="tag set: tagType:tag, comma/space separated (per type \u2264'],
    ['<input class="inp" id="docs-q" placeholder="\u6309 kind/scope/name \u8FC7\u6EE4\u2026" value="', '<input class="inp" id="docs-q" placeholder="Filter by kind/scope/name\u2026" value="'],
    ['<input class="inp" id="ev-q" placeholder="\u6309\u7C7B\u578B/\u5BF9\u8C61/\u6458\u8981\u8FC7\u6EE4\u2026" value="', '<input class="inp" id="ev-q" placeholder="Filter by type/object/summary\u2026" value="'],
    ['<section class="card obj-content"><div class="card-h"><span class="h-title">\u5185\u5BB9</span><span class="h-right">', '<section class="card obj-content"><div class="card-h"><span class="h-title">Content</span><span class="h-right">'],
    ['<section class="card"><div class="card-h"><span class="h-title">\u4E0A\u4E0B\u6587</span></div><div class="card-b">', '<section class="card"><div class="card-h"><span class="h-title">Context</span></div><div class="card-b">'],
    ['<section class="card"><div class="card-h"><span class="h-title">\u64CD\u4F5C</span></div><div class="card-b">', '<section class="card"><div class="card-h"><span class="h-title">Actions</span></div><div class="card-b">'],
    ['<section class="card"><div class="card-h"><span class="h-title">\u6807\u7B7E</span></div><div class="card-b" data-panel="tags">', '<section class="card"><div class="card-h"><span class="h-title">Tags</span></div><div class="card-b" data-panel="tags">'],
    ['<section class="card"><div class="card-h"><span class="h-title">\u9002\u7528\u89C4\u5219</span></div><div class="card-b" data-panel="rules">', '<section class="card"><div class="card-h"><span class="h-title">Applicable rules</span></div><div class="card-b" data-panel="rules">'],
    ['<span class="chip ok tiny" data-ed-draft-chip>\u65E0\u8349\u7A3F</span>', '<span class="chip ok tiny" data-ed-draft-chip>No draft</span>'],
    ['<span class="faint small">\u2026\u5171 ', '<span class="faint small">\u2026total '],
    ['<span class="faint small">\u672A\u88C5\u914D\uFF08tagSigSource=', '<span class="faint small">not assembled (tagSigSource='],
    ['<span class="faint">\u8981\u7D20\u5B57\u6BB5\u7F3A\u5E2D</span>', '<span class="faint">Essential fields absent \u2014 never fabricated</span>'],
    ['<span class="faint">\uFF08\u7A7A\uFF09</span>', '<span class="faint">(empty)</span>'],
    ['<span class="mono faint small">\uFF08\u5D4C\u5957\u7ED3\u6784\u89C1\u300CYAML\u300D\uFF09</span>', '<span class="mono faint small">(nested structures in "YAML")</span>'],
    ['<span class="small faint" data-conn-label>\u8FDE\u63A5\u4E2D\u2026</span></div></aside>', '<span class="small faint" data-conn-label>Connecting\u2026</span></div></aside>'],
    ['<span class="small faint">\u5FAA\u73AF\u72B6\u6001\u6682\u4E0D\u53EF\u8FBE\uFF08', '<span class="small faint">Loop status unreachable ('],
    ['<span class="small faint">\u672C\u5468\u671F\u6682\u65E0\u5FAA\u73AF\u8BB0\u5F55\u3002\u70B9\u706B\u4E0B\u4E00\u8F6E\uFF1A<code class="mono">corectl loop</code></span></div></div></div>', '<span class="small faint">No loop records this cycle. Ignite the next round: <code class="mono">corectl loop</code></span></div></div></div>'],
    ['<span class="small faint">\u8BC1\u636E\uFF1A\u2014</span>', '<span class="small faint">Evidence: \u2014</span>'],
    ['<span><span class="logo-name">Core Console</span><br><span class="logo-sub">web \u91CD\u6784\u8D77\u70B9</span></span></div>', '<span><span class="logo-name">Core Console</span><br><span class="logo-sub">web rebuild baseline</span></span></div>'],
    ['<textarea class="doc-editor" data-ed-textarea spellcheck="false" placeholder="\u539F\u6587\u7F16\u8F91\u2014\u2014\u53D8\u66F4\u5373\u7D2F\u79EF\u8349\u7A3F"', '<textarea class="doc-editor" data-ed-textarea spellcheck="false" placeholder="Edit source \u2014 changes accumulate a draft"'],
    ["<th>\u65F6\u523B</th><th>\u5224\u5B9A</th><th>\u5BF9\u8C61</th><th>\u5224\u5B9A\u4E3B\u4F53</th><th>rv</th></tr></thead><tbody>", "<th>Time</th><th>Decision</th><th>Object</th><th>Decided by</th><th>rv</th></tr></thead><tbody>"],
    [">\u64A4\u9500\u4E0A\u6B21</button>", ">Undo last</button>"],
    ["Apply \u5931\u8D25\uFF08", "Apply failed ("],
    ["Apply \u6210\u529F\u2014\u2014\u6587\u6863\u5DF2\u843D\u5E93\uFF08rv=", "Apply succeeded \u2014 document persisted (rv="],
    ["Apply \u843D\u5E93 ", "Apply persisted "],
    ["Apply\uFF08\u65E0\u53D8\u66F4\uFF09", "Apply (no changes)"],
    ["B vs \u843D\u5E93\u503C ", "B vs persisted "],
    ["B\uFF08\u73B0\u8BFB\uFF0C\u4E0D\u7F13\u5B58\uFF09 \xB7 ", "B (read live, not cached) \xB7 "],
    ["ConnectorOps / WorkService / Provider / Ref \u56DB\u65CF\uFF08\u6587\u6863\u4E0E\u8D44\u6E90\u540C\u89C4\u5219\uFF09", "ConnectorOps / WorkService / Provider / Ref families (same rules for docs and resources)"],
    ["\u5F53\u524D\u6CA1\u6709\u9700\u8981\u5173\u6CE8\u7684\u98CE\u9669\u9879\u3002", "attention aggregate zero hits (explicit empty state \u2014 never fakes risk)."],
    ["dry-run \u9884\u68C0\u672A\u8FC7\uFF08", "dry-run precheck failed ("],
    ["kinds \u2192 scope \u2192 \u5BF9\u8C61\uFF08\u5FEB\u7167\u805A\u5408\u6295\u5F71\uFF1B\u5BF9\u8C61\u540D\u53EF\u70B9\u8FDB\u5DE5\u4F5C\u53F0\uFF09", "kinds \u2192 scope \u2192 objects (snapshot aggregate projection; object names click into the workbench)"],
    ["severity \u6392\u5E8F \xB7 \u8BC1\u636E\u53EF\u70B9 \xB7 \u6570\u636E\u53EF\u4FE1\u5EA6\u663E\u5F0F", "severity-sorted \xB7 clickable evidence \xB7 explicit data confidence"],
    ["takeBudget \u987B\u4E3A ", "takeBudget must be "],
    ["workflow \u951A ", "workflow anchor "],
    ["\u2014\u2014contextidx \u672A\u88C5\u914D\uFF08\u9632\u5FA1\u9762\uFF09", " \u2014 contextidx not assembled (defensive face)"],
    ["\u2014\u2014\u4E0D\u5192\u5145\u5728\u518C\u4F1A\u8BDD\u3002</div></div></div>", " \u2014 never pretends to be a registered session.</div></div></div>"],
    ["\u2014\u2014\u7248\u672C\u51B2\u7A81\uFF1A\u7A3F\u5DF2\u4FDD\u7559\uFF0C\u8BF7\u5BF9\u9F50\u6700\u65B0\u7248\u672C\u540E\u91CD\u8BD5", " \u2014 optimistic-lock conflict: draft preserved; align to the latest rv and retry"],
    ["\u2014\u2014\u7248\u672C\u51B2\u7A81\uFF1A\u8BF7\u5237\u65B0\u9762\u677F\u5BF9\u9F50\u6700\u65B0\u7248\u672C\u540E\u91CD\u8BD5\uFF08\u539F\u7A3F\u4E0D\u8986\u76D6\uFF09", " \u2014 optimistic-lock conflict: reopen the panel to align revision and retry (draft not overwritten, no silent rebase)"],
    ["\u2014\u2014\u843D\u5E93\u524D\u8BF7\u5148\u5237\u65B0\uFF09", " \u2014 persistence would be rejected)"],
    ["\u2190 \u4EFB\u52A1\u4E2D\u5FC3", "\u2190 Task Center"],
    ["\u2190 \u4F1A\u8BDD", "\u2190 Sessions"],
    ["\u2190 \u5DE5\u4F5C\u6D41", "\u2190 Workflows"],
    ["\u25CF \u672A Apply\uFF08\u8349\u7A3F ", "\u25CF Not applied (draft "],
    ["\u300D\u540C\u7EC4", '" group'],
    ["\u4E00\u822C", "medium"],
    ["\u4E0D\u88C5\u5047\u2014\u2014scope \u8BCD\u5F62\u53EF\u80FD\u4E3A\u5B9E\u4F8B\u6876\uFF08\u6811\u5185\u53EF\u89C1\uFF09\u3002", "Never faked \u2014 scope word form may be an instance bucket (visible in the tree)."],
    ["\u4E25\u91CD", "critical"],
    ["\u7248\u672C\u5BF9\u9F50 r", "optimistic-lock align r"],
    ["\u7248\u672C\u5DF2\u53D8\u5316\uFF08\u73B0 r", "optimistic-lock drift (now r"],
    ["\u4E8B\u4EF6", "Events"],
    ["\u4EFB\u52A1\u4E0D\u5728\u518C", "Task not registered"],
    ["\u4F1A\u8BDD ", "Session "],
    ["\u4F1A\u8BDD\u539F\u6587\u4E0D\u5165\u516C\u5171\u5E93\uFF08intake \u7559\u75D5\u9762\u4E3A\u51C6\uFF09\u3002</div>", "Raw session text stays out of the public library (intake ledger face is authoritative).</div>"],
    ["\u4F1A\u8BDD\u8BE6\u60C5", "Session detail"],
    ["\u4F1A\u8BDD\u8F6C\u5F55", "Session transcript"],
    ["\u4F5C\u7528\u57DF", "Scope"],
    ["\u4F9D\u8D56", "Depends on"],
    ["\u4F9D\u8D56 ", "Depends on "],
    ["\u516D\u7247\u56FA\u5B9A\uFF1A\u6807\u7B7E/\u89C4\u5219/\u4E0A\u4E0B\u6587/\u64CD\u4F5C/\u5185\u5BB9/scheme", "Six fixed panels: tags/rules/context/actions/content/scheme"],
    ["\u5171 ", "total "],
    ["\u5173\u8054\u4F1A\u8BDD", "Related sessions"],
    ["\u5173\u8054\u8FD0\u884C", "Related runs"],
    ["\u5185\u7F6E\u670D\u52A1", "Built-in services"],
    ["\u51ED\u636E\u6E90", "Credential source"],
    ["\u539F\u6587", "Source"],
    ["\u539F\u6587\uFF08HEAD\uFF09", "Source (HEAD)"],
    ["\u53BB\u5904\u7406", "Handle"],
    ["\u53BB\u5BF9\u8C61\u5DE5\u4F5C\u53F0", "Go to Object Workbench"],
    ["\u53D7\u963B", "Blocked"],
    ["\u53EF\u64A4\u9500", "revertible"],
    ["\u540C\u540D\u5BF9\u8C61\u8DE8 kind/scope \u4E0D\u552F\u4E00\u2014\u2014\u8BF7\u4ECE\u7CFB\u7EDF\u6811\u8FDB\u5165\u3002", "Object name is ambiguous across kinds/scopes \u2014 enter via the System Tree."],
    ["\u540D\u79F0", "Name"],
    ["\u56DE Welcome", "Back to Welcome"],
    ["\u5728\u6587\u6863\u5E93\u4E2D\u6253\u5F00", "Open in Docs"],
    ["\u57FA\u51C6=r", "base=r"],
    ["\u57FA\u51C6=r?", "base=r?"],
    ["\u57FA\u7840\u6C47\u603B", "Summary"],
    ["\u5916\u90E8\u4F9D\u8D56", "External dependencies"],
    ["\u5931\u8D25", "failed"],
    ["\u5BF9\u8BDD\u5165\u53E3\uFF08Agent\uFF09", "Chat entries (Agent)"],
    ["\u5BF9\u8C61\u5DE5\u4F5C\u53F0\uFF08\u516D\u7247\uFF1A\u6807\u7B7E/\u89C4\u5219/\u4E0A\u4E0B\u6587/\u64CD\u4F5C/\u5185\u5BB9/scheme\uFF09", "Object Workbench (six panels: tags/rules/context/actions/content/scheme)"],
    ["\u5BF9\u8C61\u6B63\u6587\u5F15\u7528 ", "Object content reference "],
    ["\u5BF9\u8C61\u6B63\u6587\u6309 resourceVersion \u7F13\u5B58\uFF08rv \u672A\u53D8\u4E0D\u91CD\u8BFB\uFF09\u3002</div>", "Object content cached by resourceVersion (no re-read while rv unchanged).</div>"],
    ["\u5BF9\u8C61\u89E3\u6790\u6B67\u4E49\u6216\u7F3A\u5E2D", "Object resolution ambiguous or absent"],
    ["\u5DE5\u4F5C\u6D41\u4E0D\u5728\u518C", "Workflow not registered"],
    ["\u5DE5\u4F5C\u6D41\u8BE6\u60C5", "Workflow detail"],
    ["\u5DE6\u4FA7\u9009\u62E9\u4ED3\u6587\u4EF6\uFF08\u9ED8\u8BA4\u9996\u4ED3\u9996\u6587\u4EF6\uFF09\u3002", "Pick a repo file on the left (defaults to first repo, first file)."],
    ["\u5DF2\u540C\u6B65", "in snapshot"],
    ["\u5DF2\u53D6\u6D88", "canceled"],
    ["\u5DF2\u5B8C\u6210", "done"],
    ["\u5DF2\u5E94\u7528\u7684\u7248\u672C", "Applied version"],
    ["\u5DF2\u6302\u8F7D\u8FDE\u63A5", "Mounted connectors"],
    ["\u5F53\u524D\u4F1A\u8BDD", "Current session"],
    ["\u5F53\u524D\u5FAA\u73AF", "Current loop"],
    ["\u5F53\u524D\u65E0\u8FDB\u884C\u4E2D\u5FAA\u73AF", "No loop in progress"],
    ["\u5F53\u524D\u8FC7\u6EE4\u96F6\u547D\u4E2D\u2014\u2014\u653E\u5BBD\u6761\u4EF6\u518D\u8BD5\u3002", "Filter has zero hits \u2014 relax the conditions and retry."],
    ["\u5F85\u542F\u52A8", "pending start"],
    ["\u5F85\u5B9A", "pending"],
    ["\u5F85\u5BA1\u533A=\u8349\u7A3F\u7F13\u5B58\uFF08DocReview \u7C7B\u540C\u6E90\uFF09\u222A ApprovalNeeded \u5F52\u6863\u884C\uFF08\u5DF2\u6279\u7684\u4E0D\u51FA\u73B0\uFF09\u3002\u6279\u51C6/\u9A73\u56DE\u5747=\u4EBA\u70B9\uFF1Bdiff=\u7A3F vs \u6587\u6863\u73B0\u503C\u73B0\u8BFB\u3002", "Pending = draft cache (DocReview same source) \u222A ApprovalNeeded archive rows (approved ones drop out). Approve/reject are human clicks; diff = draft vs live document, read fresh."],
    ["\u6682\u65E0\u5F85\u5BA1\u4E8B\u9879", "Pending is empty (both draft cache and ApprovalNeeded archive are empty)"],
    ["\u5FAA\u73AF\u5524\u9192 #", "Loop wake #"],
    ["\u5FAA\u73AF\u72B6\u6001\u73B0\u8BFB\u4E2D\u2026", "Loading loop status\u2026"],
    ["\u5FEB\u7167", "Snapshot"],
    ["\u5FEB\u7167 events \u805A\u5408\u4E3A\u7A7A\uFF08\u4E0D\u88C5\u5047\uFF09\u3002", "Snapshot events aggregate is empty (never faked)."],
    ["\u6682\u65E0\u5728\u518C\u5BF9\u8C61\u3002", "Snapshot objects aggregate empty or no objects for this kind (never faked)."],
    ["\u5FEB\u7167 proposals \u805A\u5408\u4E3A\u7A7A\u3002", "Snapshot proposals aggregate is empty (explicit empty state)."],
    ["\u6682\u65E0\u6B64\u8FD0\u884C\u8BB0\u5F55\uFF08\u53EF\u80FD\u5DF2\u5F52\u6863\uFF09\u3002", "Snapshot runs aggregate has no such run (may have left the window)."],
    ["\u5FEB\u7167 workflows \u805A\u5408\u4E3A\u7A7A\uFF08\u4E0D\u88C5\u5047\uFF09\u3002", "Snapshot workflows aggregate is empty (never faked)."],
    ["\u6682\u65E0\u6B64\u4EFB\u52A1\uFF08\u53EF\u80FD\u5DF2\u5B8C\u6210\u5F52\u6863\u6216 id \u6709\u8BEF\uFF09\u3002", "No such id in the snapshot task list (may be done and out of window, or a wrong id)."],
    ["\u6682\u65E0\u5BF9\u8C61", "Snapshot carries no object enumeration"],
    ["\u6682\u65E0\u6B64\u5DE5\u4F5C\u6D41\uFF08\u53EF\u80FD\u5DF2\u4E0B\u7EBF\u6216\u540D\u79F0\u6709\u8BEF\uFF09\u3002", "Snapshot aggregate has no such workflow (may be offline or a wrong name)."],
    ["\u6253\u5F00\u7CFB\u7EDF\u6811", "Open System Tree"],
    ["\u6279\u51C6\uFF08= Apply\uFF09", "Approve (= Apply)"],
    ["\u6279\u91CF\u6279\u51C6\u901A\u8DEF\u5F85\u540E\u7AEF\u8BCD\u5F62\u2014\u2014\u9010\u6761\u6279\u51C6\uFF08\u8BDA\u5B9E\u6CE8\u8BB0\uFF09", "Batch approval awaits a backend word form \u2014 approve one by one (honest note)"],
    ["\u6392\u961F\u4E2D", "queued"],
    ["\u63D0\u6848\u9762\uFF08\u5217\u8868+\u8981\u7D20\u73B0\u8BFB\uFF09\u2014\u2014\u6279\u51C6/\u9A73\u56DE\u8D70\u5BA1\u6279\u4E0E Apply\uFF08LLM \u65E0 Apply \u6743\uFF09", "Proposal face (list + live essentials) \u2014 approve/reject go through Approvals and Apply (the LLM has no Apply right)"],
    ["\u64A4\u9500\u6700\u8FD1\u4E00\u6B21\u4E0A\u4E0B\u6587\u5206\u53C9\uFF08\u6062\u590D\u670D\u52A1\u7AEF\u5199\u524D\u5FEB\u7167\uFF09\uFF1F", "Revert the latest context fork (restore the server-side pre-write snapshot)?"],
    ["\u6587\u6863\u4E0E\u8D44\u6E90\u540C\u89C4\u5219\u2014\u2014\u540C\u4E00\u6E32\u67D3\u6838\uFF1B\u25CF=\u6709\u672A Apply \u8349\u7A3F", "Docs and resources share the rules \u2014 one render core; \u25CF = unapplied draft exists"],
    ["\u6587\u6863\u7F16\u8F91\u5668", "Document editor"],
    ["\u6587\u6863\u7F3A\u5E2D\uFF08\u5728\u518C\u7A3F\u57FA\u5EA7\u2014\u2014Apply=\u521B\u5EFA\u901A\u8DEF\uFF09", "Document absent (registered draft base \u2014 Apply = create path)"],
    ["\u65E0\u7A3F", "no draft"],
    ["\u65E0\u8349\u7A3F", "No draft"],
    ["\u65F6\u95F4\u672A\u77E5", "time unknown"],
    ["\u667A\u80FD\u5F52\u7EB3", "Smart summary"],
    ["\u6682\u65E0\u5728\u518C\u5DE5\u4F5C\u6D41", "No registered workflows"],
    ["\u6682\u65E0\u5F85\u529E", "No to-dos"],
    ["\u6682\u65E0\u63D0\u6848", "No proposals"],
    ["\u6700\u8FD1\u52A8\u4F5C", "Last action"],
    ["\u6700\u8FD1\u52A8\u4F5C\uFF1A\u2014", "Last action: \u2014"],
    ["\u6700\u8FD1\u9519\u8BEF", "Last error"],
    ["\u670D\u52A1\u5B9E\u4F8B", "Service instances"],
    ["\u672A\u5206\u7EA7", "unranked"],
    ["\u672A\u5206\u7EC4", "Ungrouped"],
    ["\u672A\u547D\u540D\u4E8B\u4EF6", "Unnamed event"],
    ["\u672A\u547D\u540D\u4EFB\u52A1", "Unnamed task"],
    ["\u672A\u547D\u540D\u5173\u6CE8\u9879", "Unnamed attention item"],
    ["\u672A\u547D\u540D\u63D0\u6848", "Unnamed proposal"],
    ["\u672A\u6CE8\u5165=\u6E05\u5355\u964D\u7EA7\u5FEB\u7167\u6295\u5F71\uFF0C\u8F6C\u5F55\u9762\u663E\u5F0F\u7F3A\u5E2D\u8BCD\u5F62\u3002</div>", "Not injected = list degrades to snapshot projection; the transcript face is explicitly absent.</div>"],
    ["\u672A\u6CE8\u518C\u58F0\u660E kind\u2014\u2014\u5168\u91CF\u679A\u4E3E\u515C\u5E95", "Unregistered declared kind \u2014 falls back to full enumeration"],
    ["\u751F\u6210\u65B9\u5F0F\u672A\u6CE8\u660E", "generation method unspecified"],
    ["\u672A\u77E5", "unknown"],
    ["\u672A\u77E5\u8584\u5C4F", "Unknown thin screen"],
    ["\u672A\u9009\u6863", "No file selected"],
    ["\u672C\u7A97\u53E3\u65E0\u5173\u6CE8\u9879", "No attention items in this window"],
    ["\u67E5\u770B ", "View "],
    ["\u6807\u7B7E\u300C", 'Tag "'],
    ["\u6807\u9898", "Title"],
    ["\u6B63\u6587\u6682\u4E0D\u53EF\u8FBE\uFF08", "Content unreachable ("],
    ["\u52A0\u8F7D\u4E2D\u2026", "Hydrating\u2026"],
    ["\u7248\u672C\u53F7", "Version"],
    ["\u72B6\u6001", "Status"],
    ["\u73B0\u8BFB\u4E2D\u2026", "Loading\u2026"],
    ["\u767D\u540D\u5355\u4ED3\u6811\uFF08GET /v1/design-docs\u2014\u2014HEAD \u6743\u5A01\u96F6\u7F13\u5B58\uFF09", "Whitelisted repo tree (GET /v1/design-docs \u2014 HEAD authoritative, zero cache)"],
    ["\u76EE\u6807", "Target"],
    ["\u76EE\u6807\u6587\u6863", "Target document"],
    ["\u76F8\u5173\u6587\u6863", "Related documents"],
    ["\u770B root \u4F1A\u8BDD", "View root session"],
    ["\u786E\u8BA4\u6279\u51C6\uFF08Apply\uFF09", "Confirm approve (Apply)"],
    ["\u7A3F: ", "Draft: "],
    ["\u7A97\u53E3\u5185\u65E0\u4E8B\u4EF6", "No events in window"],
    ["\u6682\u65E0\u6587\u6863", "No documents in window"],
    ["\u7B49\u5F85\u9996\u4E2A\u5FEB\u7167\u805A\u5408\u3002", "Waiting for the first snapshot aggregate."],
    ["\u7B49\u5F85\u9996\u4E2A\u5FEB\u7167\u805A\u5408\u6216 patch \u589E\u91CF", "Waiting for first snapshot aggregate or patch increment"],
    ["\u7ED3\u675F\u4E8E", "ended at"],
    ["\u7F51\u7EDC", "Network"],
    ["\u805A\u5408\u9762\uFF08\u5FEB\u7167 workflows/runs\uFF09\uFF1B\u8FD0\u884C\u72EC\u7ACB\u9875\u89C1\u8BE6\u60C5", "Aggregate face (snapshot workflows/runs); runs have their own detail page"],
    ["\u8349\u7A3F\u222A\u5BA1\u6279\u4E8B\u4EF6\u2014\u2014\u70B9\u51FB\u5904\u7406", "Drafts \u222A approval events \u2014 click to handle"],
    ["\u8349\u7A3F\u4FDD\u5B58\u5931\u8D25\uFF08", "Draft save failed ("],
    ["\u843D\u5E93\u4E2D\u2026", "Persisting\u2026"],
    ["\u89C4\u5219\u539F\u6587\u63D0\u53CA ", "Rule source mentions "],
    ["\u89C4\u5219\u6587\u6863\u65CF\u7D22\u5F15\u6682\u4E0D\u53EF\u8FBE\u2014\u2014\u547D\u4E2D\u5224\u5B9A\u6682\u7F3A\uFF08\u4E0D\u5192\u5145\u65E0\u9002\u7528\uFF09\u3002", "Rule doc family index unreachable \u2014 hit evaluation unavailable (never pretends none apply)."],
    ["\u89E6\u53D1\u4E8B\u4EF6", "Trigger event"],
    ["\u8BC1\u636E", "Evidence"],
    ["\u8BC1\u636E\u2014\u2014\u70B9\u51FB\u8FDB\u5165", "Evidence \u2014 click to open"],
    ["\u8BE5 scope \u6682\u65E0\u5728\u518C\u5BF9\u8C61", "No registered objects for this scope"],
    ["\u8BF4\u660E\u6587\u6863", "Guide"],
    ["\u8D1F\u8D23 ", "Owner "],
    ["\u8FC7\u6EE4\u96F6\u547D\u4E2D\u6216\u5FEB\u7167\u4E3A\u7A7A\u3002", "Filter zero hits or snapshot empty (explicit empty state)."],
    ["\u8FD0\u884C\u72B6\u6001", "Run status"],
    ["\u8FD0\u884C\u8BB0\u5F55\u4E0D\u5728\u518C", "Run record not registered"],
    ["\u8FD0\u884C\u8BE6\u60C5", "Run detail"],
    ["\u91CD\u8981", "important"],
    ["\u9519\u8BEF", "error"],
    ["\u9700\u6C42\u63D0\u6848", "Requirement proposal"],
    ["\u9884\u68C0\u4E2D\u2026", "Prechecking\u2026"],
    ["\u9A8C\u8BC1\u4E2D", "verifying"],
    ["\uFF08faces \u672A\u6CE8\u5165\u4E14\u5FEB\u7167 sessions \u7F3A\u5E2D\uFF09", "(faces not injected and snapshot sessions absent)"],
    ["\uFF08git \u6743\u5A01\u96F6\u7F13\u5B58\uFF09 \xB7 \u8349\u7A3F/Apply \u901A\u8DEF\u7ECF\u7F16\u8F91\u5668\u5C4F\uFF08doc/\u2026\uFF09\u3002</div>", "(git authoritative, zero cache) \xB7 draft/Apply path goes through the editor screen (doc/\u2026).</div>"],
    ["</span>", "(content not re-read while rv unchanged \u2014 applies to the degraded polling path too)</span>"],
    ["\uFF08\u7248\u672C\u57FA\u51C6\uFF09", "(optimistic-lock base)"],
    ["\uFF08\u843D\u5E93\u671F\u671B\u7248\u672C\u2014\u2014\u7248\u672C\u53D8\u5316\u65F6\u9700\u5237\u65B0\u91CD\u8BD5\uFF09\xB7 \u7A3F ", "(optimistic-lock expected \u2014 409 never silently rebases) \xB7 draft "],
    ["\uFF08\u5BF9\u8BDD\u662F\u4E00\u7B49\u9762\u2014\u2014\u540C\u4E00 dsh \u4F1A\u8BDD\u673A\u5236\u6295\u5F71\uFF09", "(chat is a first-class face \u2014 projection of the same dsh session mechanism)"],
    ["\uFF08\u5BF9\u8C61\u5C1A\u672A\u843D\u5E93\u2014\u2014\u6279\u51C6\u5373\u521B\u5EFA\uFF09", "(object not registered \u2014 apply is the create path)"],
    ["\uFF08\u65E0\u5DEE\u5F02\uFF09", "(no diff)"],
    ["\uFF08\u9762\u73B0\u503C\u6682\u4E0D\u53EF\u8BFB\uFF1A", "(face current value unreadable: "],
    ["\uFF09\u2014\u2014HEAD \u6743\u5A01\u96F6\u7F13\u5B58\uFF0C\u91CD\u8FDB\u5373\u73B0\u8BFB\u3002</div>", ") \u2014 HEAD authoritative, zero cache; re-enter to read fresh.</div>"],
    ["\uFF09\u2014\u2014\u4E0D\u5192\u5145\u65E0\u5FAA\u73AF\u3002</span></div></div></div>", ") \u2014 never pretends there is no loop.</span></div></div></div>"],
    ["\uFF09\u3002</div>", ") \u2014 decision trail is currently apply_log only.</div>"],
    ["\uFF09\u2014\u2014\u5DE5\u4F5C\u526F\u672C\u4ECD\u5728\u5185\u5B58\uFF0C\u9000\u51FA\u524D\u4E0D\u4E22", ") \u2014 the working copy stays in memory, not lost until you leave"],
    ["\uFF09\u2014\u2014\u6062\u590D\u540E\u91CD\u8FDB\u5373\u73B0\u8BFB", ") \u2014 re-enter after recovery to read fresh"],
    ["\uFF09\u2014\u2014\u6062\u590D\u540E\u91CD\u8FDB\u5373\u73B0\u8BFB\u3002</div>", ") \u2014 re-enter after recovery to read fresh.</div>"],
    ["\uFF09\u2014\u2014\u6808\u7A7A\u4E0D\u53EF\u64A4\uFF08\u4E0D\u5192\u5145\u64A4\u9500\u6210\u529F\uFF09\u3002</div>", ") \u2014 empty stack cannot revert (never pretends a successful revert).</div>"],
    ["\uFF09\u2014\u2014\u7A3F\u4FDD\u7559\u8BED\u4E49\uFF0C\u672A\u843D\u5E93\u3002", ") \u2014 draft-preserved semantics, not persisted."],
    ["\uFF09\u2014\u2014\u91CD\u8FDB\u5373\u6062\u590D\u3002</div>", ") \u2014 re-enter to restore.</div>"],
    ["\uFF09\u2014\u2014\u96F6\u526F\u4F5C\u7528\uFF0C\u7A3F\u5DF2\u4FDD\u7559\u3002", ") \u2014 zero side effects, draft preserved."],
    ["\uFF09\u3002\u6062\u590D\u540E\u91CD\u8FDB\u5373\u73B0\u8BFB\u3002</div></div></div>", "). Re-enter after recovery to read fresh.</div></div></div>"],
    ["\uFF09\uFF1F\u843D\u5E93\u7248\u672C\u57FA\u51C6 r", ")? optimistic-lock base r"],
    ["\uFF1A\u6682\u65E0\u3002</div>", ": none (explicit empty state).</div>"],
    ["\uFF1B\u8D85\u9650 tag \u6536\u655B ", "; over-limit tags converge "],
    ["\uFF1Fdry-run \u5DF2\u8FC7\uFF08rv=", "? dry-run passed (rv="],
    /* §19 全页可用性批——EN 回填（plain 缺译全量） */
    ["\u6587\u6863\u7F16\u8F91\u5668", "Document editor"],
    ["\u8FD0\u884C\u8BE6\u60C5", "Run detail"],
    ["Service \u5927\u5C4F", "Service dashboard"],
    ["\u7CFB\u7EDF\u8BE6\u60C5", "System detail"],
    ["\uFF08\u65E0\u5DEE\u5F02\uFF09", "(no diff)"],
    ["\u540D\u79F0", "Name"],
    ["\u4F5C\u7528\u57DF", "Scope"],
    ["\u7248\u672C\u53F7", "Version"],
    ["\u76EE\u6807\u6587\u6863", "Target doc"],
    ["\u8BF4\u660E\u6587\u6863", "Description"],
    ["\u76F8\u5173\u6587\u6863", "Related docs"],
    ["\u5916\u90E8\u4F9D\u8D56", "External deps"],
    ["\u670D\u52A1\u5B9E\u4F8B", "Service instances"],
    ["\u5185\u7F6E\u670D\u52A1", "Built-in services"],
    ["\u5BF9\u8BDD\u5165\u53E3\uFF08Agent\uFF09", "Dialog entry (Agent)"],
    ["\u8FD0\u884C\u72B6\u6001", "Run status"],
    ["\u5DF2\u5E94\u7528\u7684\u7248\u672C", "Applied version"],
    ["\u5DF2\u6302\u8F7D\u8FDE\u63A5", "Mounted connections"],
    ["\u51ED\u636E\u6E90", "Credential source"],
    ["\u5F53\u524D\u4F1A\u8BDD", "Current session"],
    ["\u6807\u9898", "Title"],
    ["\u4F9D\u8D56", "Deps"],
    ["\u539F\u6587", "Source"],
    ["\u67E5\u770B\u5BF9\u8C61", "View object"],
    ["\u5C55\u5F00", "Expand"],
    ["\u6211", "Me"],
    ["\u751F\u6210\u7ED3\u8BBA", "Conclusion"],
    ["\u5B8C\u6574\u6587\u6863\u6A21\u5F0F\uFF1A", "Full document: "],
    ["\u5B8C\u6574\u6587\u6863\u6A21\u5F0F", "Full document"],
    ["\u5199\u5165\u8349\u7A3F", "To draft"],
    ["\u601D\u8003\u8FC7\u7A0B", "Thinking"],
    ["\u5DE5\u5177\u8FC7\u7A0B", "Tool trace"],
    ["\u53BB\u5BF9\u8C61\u5DE5\u4F5C\u53F0", "Object workspace"],
    ["\u7A3F: ", "Draft: "],
    ["\uFF08\u5BF9\u8C61\u5C1A\u672A\u843D\u5E93\u2014\u2014\u6279\u51C6\u5373\u521B\u5EFA\uFF09", "(object not registered \u2014 apply will create it)"],
    ["\u57FA\u51C6=r?", "base=r?"],
    ["\u57FA\u51C6=r", "base=r"],
    ["\uFF08\u9762\u73B0\u503C\u6682\u4E0D\u53EF\u8BFB\uFF1A", "(face value unreadable: "],
    [" \xB7 \u7A3F ", " \xB7 draft "],
    ["B vs \u843D\u5E93\u503C ", "B vs stored "],
    ["B\uFF08\u73B0\u8BFB\uFF0C\u4E0D\u7F13\u5B58\uFF09 \xB7 ", "B (live read, no cache) \xB7 "],
    ["\u9884\u68C0\u4E2D\u2026", "Dry-run\u2026"],
    ["\u6279\u51C6\uFF08= Apply\uFF09", "Approve (= Apply)"],
    ["dry-run \u9884\u68C0\u672A\u8FC7\uFF08", "dry-run precheck failed ("],
    ["\uFF09\u2014\u2014\u96F6\u526F\u4F5C\u7528\uFF0C\u7A3F\u5DF2\u4FDD\u7559\u3002", ") \u2014 zero side effects, draft kept."],
    ["\u7248\u672C\u5BF9\u9F50 r", "optimistic lock aligned r"],
    ["\u7248\u672C\u5DF2\u53D8\u5316\uFF08\u73B0 r", "optimistic lock drift (now r"],
    [" vs \u7F16\u8F91\u57FA\u7EBF r", " vs draft base r"],
    ["\u2014\u2014\u843D\u5E93\u524D\u8BF7\u5148\u5237\u65B0\uFF09", " \u2014 apply would be rejected)"],
    ["\u786E\u8BA4\u6279\u51C6\uFF08Apply\uFF09", "Confirm approve (Apply) "],
    ["\uFF1Fdry-run \u5DF2\u8FC7\uFF08rv=", "? dry-run passed (rv="],
    ["\u843D\u5E93\u4E2D\u2026", "Applying\u2026"],
    ["Apply \u5931\u8D25\uFF08", "Apply failed ("],
    ["\uFF09\u2014\u2014\u7A3F\u4FDD\u7559\u8BED\u4E49\uFF0C\u672A\u843D\u5E93\u3002", ") \u2014 draft kept, not applied."],
    ["\u53BB\u5904\u7406", "Handle"],
    ["\u5F85\u5BA1\u533A=\u8349\u7A3F\u7F13\u5B58\uFF08DocReview \u7C7B\u540C\u6E90\uFF09\u222A ApprovalNeeded \u5F52\u6863\u884C\uFF08\u5DF2\u6279\u7684\u4E0D\u51FA\u73B0\uFF09\u3002\u6279\u51C6/\u9A73\u56DE\u5747=\u4EBA\u70B9\uFF1Bdiff=\u7A3F vs \u6587\u6863\u73B0\u503C\u73B0\u8BFB\u3002", "Pending = draft cache (DocReview same source) \u222A ApprovalNeeded archived rows (granted ones drop out). Approve/reject are human clicks; diff = draft vs live value."],
    ["\u6279\u91CF\u6279\u51C6\u901A\u8DEF\u5F85\u540E\u7AEF\u8BCD\u5F62\u2014\u2014\u9010\u6761\u6279\u51C6\uFF08\u8BDA\u5B9E\u6CE8\u8BB0\uFF09", "Batch approve awaits backend verb \u2014 approve one by one (honest note)"],
    ["\u6682\u65E0\u5F85\u5BA1\u4E8B\u9879", "Nothing pending (draft cache and ApprovalNeeded archive both empty)"],
    ["\u56DE Welcome", "Back to Welcome"],
    ["\u2014 \u9009\u62E9 session\uFF08\u6216\u8F93\u5165\uFF09\u2014", "\u2014 pick a session (or type) \u2014"],
    ["\u6216\u624B\u586B session id", "or type a session id"],
    ["\u63D0\u4EA4\u9700\u6C42\uFF0F\u627E\u6587\u6863\uFF0F\u66F4\u65B0\u6587\u6863\u2014\u2014\u4F8B\u5982\uFF1A\u8FDB\u884C\u4EE3\u7801\u5BA1\u6838\uFF0C\u770B\u770B\u5B9E\u9645\u548C\u6587\u6863\u5DEE\u8DDD\u6709\u591A\u5927\uFF0C\u5E76\u8FDB\u884C\u4FEE\u590D", "Submit a request / find a doc / update a doc \u2014 e.g. run a code audit, measure the actual-vs-doc gap, and fix it"],
    ["\u53D1\u9001", "Send"],
    ["Enter \u53D1\u9001 \xB7 queue \u8BED\u4E49\uFF08\u9001\u8FBE agent \u4F1A\u8BDD\uFF1B\u7ED3\u8BBA\u7531 agent \u843D\u6587\u6863\u540E\u7ECF\u72B6\u6001\u5FBD\u4E0E\u8282\u70B9\u56FE\u53EF\u89C1\uFF09", "Enter to send \xB7 queue semantics (delivered to the agent session; the conclusion lands in docs by the agent, visible via status badge and node graph)"],
    ["\u5BF9\u8BDD\u901A\u8DEF\u672A\u6CE8\u5165\uFF08chatfaces.js \u7F3A\u5E2D\uFF09\u2014\u2014\u9001\u8FBE\u9762\u663E\u5F0F\u7F3A\u5E2D\uFF0C\u4E0D\u5192\u5145\u5DF2\u53D1\u9001", "Dialog channel not injected (chatfaces.js absent) \u2014 delivery face explicitly absent, never fakes a send"],
    ["\u7EBF\u7A0B\u4E3A\u7A7A\u2014\u2014\u53D1\u7B2C\u4E00\u6761\u9700\u6C42\uFF1Bagent \u7684\u5B8C\u6574\u8F68\u8FF9\u5728\u539F\u751F\u754C\u9762\uFF08\u65B0\u7A97\u53E3\uFF09\u53EF\u89C1\u3002", "Thread empty \u2014 send the first request; the agent\u2019s full track is visible in the native UI (new window)."],
    ["\u6B63\u6587\u6682\u4E0D\u53EF\u8BFB\uFF08", "Body unreadable ("],
    ["\u5185\u5BB9\u7F3A\u5E2D", "content absent"],
    ["\u5173\u95ED", "Close"],
    ["\u5199\u5165\u8349\u7A3F\uFF08\u7F16\u8F91\u5668\uFF09", "To draft (editor)"],
    ["\u5DF2\u9001\u8FBE\uFF08queue\uFF09\u2014\u2014agent \u4F1A\u8BDD\u5DF2\u53D7\u7406\u3002\u7ED3\u8BBA\u5C06\u7531 agent \u751F\u6210\u5E76\u81EA\u884C\u843D\u6587\u6863\uFF1B\u6587\u6863\u72B6\u6001\u5FBD\u4E0E workflow \u8282\u70B9\u56FE\u968F\u540E\u53EF\u89C1\u3002\u539F\u751F\u8F68\u8FF9\uFF1A", "Delivered (queue) \u2014 agent session accepted. The conclusion will be produced by the agent and land in docs on its own; watch the doc status badge and workflow node graph. Native track:"],
    ["\u9001\u8FBE\u5931\u8D25\uFF08", "Delivery failed ("],
    ["\u5BBF\u4E3B\u539F\u751F\u754C\u9762\uFF08\u6743\u5A01\u8F68\u8FF9\uFF09", "Host native UI (authoritative track)"],
    ["rootAgent\u2014\u2014\u627E\u6587\u6863\u3001\u66F4\u65B0\u6587\u6863\uFF1Bref \u53EF\u70B9\uFF0C\u7ED3\u8BBA\u53EF\u5168\u6587\u6863\u89C2\u770B", "rootAgent \u2014 find and update docs; refs are clickable, conclusions viewable as full documents"],
    ["\u65B0\u7A97\u53E3\u539F\u751F\u754C\u9762", "Native UI in new window"],
    ["\u5BF9\u8BDD\u7EBF\u7A0B", "Dialog thread"],
    ["\u672C\u5730\u7EBF\u7A0B\u53EA\u8BB0\u672C\u63A7\u5236\u53F0\u53D1\u51FA\u7684\u6D88\u606F\uFF1B\u6743\u5A01\u8F68\u8FF9=\u5BBF\u4E3B\u539F\u751F\u4F1A\u8BDD\uFF08\u4E0B\u65B9 expand \u5185\u5D4C\uFF09\u3002", "Local thread only records messages sent from this console; the authoritative track is the host native session (embedded below)."],
    ["\u4E0A\u4E0B\u6587\u52A0\u8F7D\u4E2D\u2026\uFF08GET /v1/sessions/", "Hydrating context\u2026 (GET /v1/sessions/"],
    ["\u65E0\u7A3F", "No draft"],
    ["\u6B63\u6587\u6682\u4E0D\u53EF\u8FBE\uFF08", "Body unreachable ("],
    ["\uFF09\u2014\u2014\u6062\u590D\u540E\u91CD\u8FDB\u5373\u73B0\u8BFB", ") \u2014 reopen after recovery to read live"],
    ["scheme\uFF08frontmatter \u58F0\u660E\u2014\u2014expand \u67E5\u770B\u5B8C\u6574\uFF09", "scheme (frontmatter declaration \u2014 expand for full)"],
    ["\u25CF \u672A Apply\uFF08\u8349\u7A3F ", "\u25CF unapplied (draft "],
    [" \u5904\u53D8\u66F4)", " change(s))"],
    ["\u65E0\u8349\u7A3F", "No draft"],
    ["Apply\uFF08\u65E0\u53D8\u66F4\uFF09", "Apply (no changes)"],
    ["\u6D41\u7A0B\uFF1A\u5199\u4FEE\u6539\u610F\u56FE \u2192 AI \u51FA\u5168\u6587\u5EFA\u8BAE \u2192 \u5199\u5165\u8349\u7A3F \u2192 \u53F3\u4E0B Apply\uFF08\u7248\u672C\u6821\u9A8C+\u9884\u68C0\u7167\u8D70\u2014\u2014AI \u65E0\u7279\u6743\uFF09\u3002", "Flow: write the change intent \u2192 AI proposes a full draft \u2192 to draft \u2192 Apply at bottom-right (optimistic lock + precheck as usual \u2014 AI has no privilege)."],
    ["\u4FEE\u6539\u610F\u56FE\uFF0C\u5982\uFF1A\u8865\u5145 X \u8282\uFF1B\u4FEE\u6B63 Y \u5904\u53E3\u5F84\u2026", "Change intent, e.g. add section X; fix wording at Y\u2026"],
    ["\u8BF7\u6C42 AI \u5EFA\u8BAE", "Ask AI for a proposal"],
    ["\u5EFA\u8BAE\u5C06\u73B0\u8BFB\u4E8E\u6B64\u2014\u2014\u5199\u5165\u8349\u7A3F\u540E\u8D70\u65E2\u6709 Apply \u901A\u8DEF\u3002", "The proposal will appear here \u2014 send to draft, then use the existing Apply path."],
    ["AI \u901A\u9053\u672A\u6CE8\u5165\uFF08__CC.chatFaces.docFix \u7F3A\u5E2D\uFF09\u2014\u2014\u53EF\u5728\u300C\u539F\u6587\u300D\u89C6\u56FE\u624B\u7F16\uFF0C\u6216\u7ECF\u4F1A\u8BDD\u5C4F\u5BF9\u8BDD\u540E\u4EBA\u5DE5\u8D34\u56DE\u3002", "AI channel not injected (__CC.chatFaces.docFix absent) \u2014 edit manually in Source view, or dialog on the session screen and paste back."],
    ["AI \u73B0\u8BFB\u4E2D\u2026", "AI working\u2026"],
    ["\u7A7A\u5EFA\u8BAE", "Empty proposal"],
    ["AI \u5EFA\u8BAE\u5DF2\u5230\uFF08", "AI proposal arrived ("],
    ["\u5DF2\u5199\u5165\u8349\u7A3F\u2014\u2014\u53F3\u4E0B Apply \u843D\u5E93\uFF08\u7248\u672C\u6821\u9A8C\u7167\u8D70\uFF09\u3002", "Written to draft \u2014 Apply at bottom-right (optimistic lock as usual)."],
    ["AI \u5EFA\u8BAE\u5931\u8D25\uFF08", "AI proposal failed ("],
    ["\uFF08\u843D\u5E93\u671F\u671B\u7248\u672C\u2014\u2014\u7248\u672C\u53D8\u5316\u65F6\u9700\u5237\u65B0\u91CD\u8BD5\uFF09\xB7 \u7A3F ", "(optimistic-lock expected \u2014 409 never silently rebases) \xB7 draft "],
    ["\u8349\u7A3F\u4FDD\u5B58\u5931\u8D25\uFF08", "Draft save failed ("],
    ["\uFF09\u2014\u2014\u5DE5\u4F5C\u526F\u672C\u4ECD\u5728\u5185\u5B58\uFF0C\u9000\u51FA\u524D\u4E0D\u4E22", ") \u2014 working copy stays in memory, not lost before exit"],
    ["\u6587\u6863\u7F3A\u5E2D\uFF08\u5728\u518C\u7A3F\u57FA\u5EA7\u2014\u2014Apply=\u521B\u5EFA\u901A\u8DEF\uFF09", "Doc absent (registered draft base \u2014 Apply = creation path)"],
    ["\uFF08\u7248\u672C\u57FA\u51C6\uFF09", "(optimistic-lock base)"],
    ["\u52A0\u8F7D\u4E2D\u2026", "Hydrating\u2026"],
    ["Apply \u843D\u5E93 ", "Apply stored "],
    ["\uFF09\uFF1F\u843D\u5E93\u7248\u672C\u57FA\u51C6 r ", ")? optimistic-lock base r "],
    ["\u2014\u2014\u7248\u672C\u51B2\u7A81\uFF1A\u7A3F\u5DF2\u4FDD\u7559\uFF0C\u8BF7\u5BF9\u9F50\u6700\u65B0\u7248\u672C\u540E\u91CD\u8BD5", " \u2014 optimistic-lock conflict: draft kept, align to latest rv and retry"],
    ["Apply \u6210\u529F\u2014\u2014\u6587\u6863\u5DF2\u843D\u5E93\uFF08rv=", "Applied \u2014 doc stored (rv="],
    ["\u5BF9\u8C61\u89E3\u6790\u6B67\u4E49\u6216\u7F3A\u5E2D", "Object resolution ambiguous or absent"],
    ["\u540C\u540D\u5BF9\u8C61\u8DE8 kind/scope \u4E0D\u552F\u4E00\u2014\u2014\u8BF7\u4ECE\u7CFB\u7EDF\u6811\u8FDB\u5165\u3002", "Same-named object across kinds/scopes \u2014 enter via the system tree."],
    ["\u6253\u5F00\u7CFB\u7EDF\u6811", "Open system tree"],
    ["\u67E5\u770B ", "View "],
    ["workflow \u951A ", "workflow anchor "],
    [" \u2192 \u5DE5\u4F5C\u6D41", " \u2192 workflow"],
    ["\u5173\u8054\u8FD0\u884C", "Related runs"],
    ["\u4F1A\u8BDD ", "Session "],
    ["\u4F1A\u8BDD\u8BE6\u60C5", "Session detail"],
    ["\u5FAA\u73AF\u5524\u9192 #", "Loop wake #"],
    ["\u672A\u547D\u540D\u4E8B\u4EF6", "Unnamed event"],
    ["\u7A97\u53E3\u5185\u65E0\u4E8B\u4EF6", "No events in window"],
    ["\u5F53\u524D\u8FC7\u6EE4\u96F6\u547D\u4E2D\u2014\u2014\u653E\u5BBD\u6761\u4EF6\u518D\u8BD5\u3002", "Filter has zero hits \u2014 loosen it and retry."],
    ["\u5FEB\u7167 events \u805A\u5408\u4E3A\u7A7A\uFF08\u4E0D\u88C5\u5047\uFF09\u3002", "Snapshot events aggregate is empty (no faking)."],
    ["\u72B6\u6001", "Status"],
    ["\u73B0\u8BFB\u4E2D\u2026", "Reading live\u2026"],
    ["kinds \u2192 scope \u2192 \u5BF9\u8C61\uFF08\u5FEB\u7167\u805A\u5408\u6295\u5F71\uFF1B\u5BF9\u8C61\u540D\u53EF\u70B9\u8FDB\u5DE5\u4F5C\u53F0\uFF09", "kinds \u2192 scope \u2192 objects (snapshot projection; names click into the workspace)"],
    ["\u6682\u65E0\u5BF9\u8C61", "Snapshot has no object enumeration"],
    ["\u7B49\u5F85\u9996\u4E2A\u5FEB\u7167\u805A\u5408\u3002", "Waiting for the first snapshot aggregate."],
    ["\u76EE\u5F55\u6811\u5F0F\u6587\u6863\u6E05\u5355\u2014\u2014tag \u8FC7\u6EE4+\u7248\u672C/\u91CD\u8981\u5EA6\u5217\uFF1B\u70B9\u884C\u8FDB\u6E32\u67D3\u89C6\u56FE\u4E0E AI \u4FEE\u6587\u6863", "Tree-style doc list \u2014 tag filter + version/importance columns; click a row for rendered view and AI doc-fix"],
    ["\u672A\u6CE8\u518C\u58F0\u660E kind\u2014\u2014\u5168\u91CF\u679A\u4E3E\u515C\u5E95", "No declared kinds registered \u2014 full enumeration fallback"],
    ["\u91CD\u8981\u5EA6\u2014", "importance \u2014"],
    ["\u6253\u5F00", "Open"],
    ["\u6E32\u67D3\u89C6\u56FE", "Rendered view"],
    ["\u6682\u65E0\u6587\u6863", "No docs in window"],
    ["\u8FC7\u6EE4\u96F6\u547D\u4E2D\u6216\u5FEB\u7167\u4E3A\u7A7A\u3002", "Filter zero-hit or snapshot empty (explicit empty state)."],
    ["tag/\u91CD\u8981\u5EA6\u6765\u81EA\u7F13\u5B58\u6B63\u6587\u7684 frontmatter\uFF08\u9884\u53D6\u7A97\u53E3 24 \u7BC7\u2014\u2014\u672A\u7F13\u5B58\u7BC7\u663E\u793A \u2014\uFF0C\u73B0\u5F00\u5373\u8BFB\uFF09\u3002", "Tags/importance come from cached body frontmatter (prefetch window 24 \u2014 uncached show \u2014, read live on open)."],
    ["\u7F51\u7EDC", "Network"],
    ["\u539F\u6587\uFF08HEAD\uFF09", "Source (HEAD)"],
    ["\u767D\u540D\u5355\u4ED3\u6811\uFF08GET /v1/design-docs\u2014\u2014HEAD \u6743\u5A01\u96F6\u7F13\u5B58\uFF09", "Whitelisted repo tree (GET /v1/design-docs \u2014 HEAD authoritative, zero cache)"],
    ["\u672A\u9009\u6863", "No file selected"],
    ["\u5DE6\u4FA7\u9009\u62E9\u4ED3\u6587\u4EF6\uFF08\u9ED8\u8BA4\u9996\u4ED3\u9996\u6587\u4EF6\uFF09\u3002", "Pick a repo file on the left (defaults to first repo, first file)."],
    ["ConnectorOps / WorkService / Provider / Ref \u56DB\u65CF\uFF08\u6587\u6863\u4E0E\u8D44\u6E90\u540C\u89C4\u5219\uFF09", "Four families: ConnectorOps / WorkService / Provider / Ref (docs and resources, same rules)"],
    ["\u9700\u6C42\u63D0\u6848", "Proposals"],
    ["\u63D0\u6848\u9762\uFF08\u5217\u8868+\u8981\u7D20\u73B0\u8BFB\uFF09\u2014\u2014\u6279\u51C6/\u9A73\u56DE\u8D70\u5BA1\u6279\u4E0E Apply\uFF08LLM \u65E0 Apply \u6743\uFF09", "Proposal face (list + live elements) \u2014 approve/reject go through approvals and Apply (LLM has no Apply right)"],
    ["\u672A\u547D\u540D\u63D0\u6848", "Unnamed proposal"],
    ["\u5F85\u5B9A", "Pending"],
    ["\u6682\u65E0\u63D0\u6848", "No proposals"],
    ["\u5FEB\u7167 proposals \u805A\u5408\u4E3A\u7A7A\u3002", "Snapshot proposals aggregate is empty (explicit empty state)."],
    [" \u7684\u5BF9\u8C61\u4E0E\u4E8B\u4EF6\uFF08\u5FEB\u7167\u805A\u5408\u6295\u5F71\uFF09", "\u2019s objects and events (snapshot projection)"],
    ["\u8BE5 scope \u6682\u65E0\u5728\u518C\u5BF9\u8C61", "No registered objects under this scope"],
    ["\u4E0D\u88C5\u5047\u2014\u2014scope \u8BCD\u5F62\u53EF\u80FD\u4E3A\u5B9E\u4F8B\u6876\uFF08\u6811\u5185\u53EF\u89C1\uFF09\u3002", "No faking \u2014 the scope word may be an instance bucket (visible in the tree)."],
    ["\u672A\u77E5\u8584\u5C4F", "Unknown thin screen"],
    ["\u4E1A\u52A1\u63A7\u5236\u53F0", "Business console"],
    ["\u4E1A\u52A1\u89C6\u89D2\u5360\u4F4D\u2014\u2014\u540E\u7EED ServiceAgent \u767B\u5F55\u540E\u5728\u6B64\u627F\u8F7D\u4E1A\u52A1\u5927\u5C4F\u4E0E\u64CD\u4F5C\u9762\uFF08\u5F53\u524D\u96F6\u5185\u5BB9\uFF0C\u663E\u5F0F\u5360\u4F4D\u4E0D\u88C5\u5047\uFF09\u3002", "Business-view placeholder \u2014 a later ServiceAgent login will host the business dashboard here (zero content now; explicit placeholder, no faking)."],
    ["\u672A\u5206\u7EC4", "Ungrouped"],
    ["\u89C4\u5219\u6587\u6863\u65CF\u7D22\u5F15\u6682\u4E0D\u53EF\u8FBE\u2014\u2014\u547D\u4E2D\u5224\u5B9A\u6682\u7F3A\uFF08\u4E0D\u5192\u5145\u65E0\u9002\u7528\uFF09\u3002", "Rule-doc family index unreachable \u2014 hit decision missing (never fakes \u201Cno applicable rules\u201D)."],
    ["\u6807\u7B7E\u300C", "Tag \u201C"],
    ["\u300D\u540C\u7EC4", "\u201D same group"],
    ["\u89C4\u5219\u539F\u6587\u63D0\u53CA ", "Rule text mentions "],
    ["\u5BF9\u8C61\u6B63\u6587\u5F15\u7528 ", "Object body references "],
    ["\u5DF2\u540C\u6B65", "In snapshot"],
    ["\u5728\u6587\u6863\u5E93\u4E2D\u6253\u5F00", "Open in doc library"],
    [" \u7F3A", " missing"],
    ["\u5BF9\u8C61\u5DE5\u4F5C\u53F0\uFF08\u516D\u7247\uFF1A\u6807\u7B7E/\u89C4\u5219/\u4E0A\u4E0B\u6587/\u64CD\u4F5C/\u5185\u5BB9/scheme\uFF09", "Object workspace (six panes: tags/rules/context/actions/content/scheme)"],
    ["\u5FEB\u7167", "Snapshot"],
    ["\u516D\u7247\u56FA\u5B9A\uFF1A\u6807\u7B7E/\u89C4\u5219/\u4E0A\u4E0B\u6587/\u64CD\u4F5C/\u5185\u5BB9/scheme", "Six fixed panes: tags/rules/context/actions/content/scheme"],
    ["\u6682\u65E0\u5728\u518C\u5BF9\u8C61\u3002", "Snapshot objects aggregate empty or no objects of this kind (no faking)."],
    ["takeBudget \u987B\u4E3A ", "takeBudget must be "],
    [" \u7684\u6570\u503C\uFF08ContextPolicy \u9884\u7B97\u57DF\uFF09", " (ContextPolicy budget domain)"],
    [" \u6B21", " time(s)"],
    [" \xB7 \u6700\u8FD1 ", " \xB7 recent "],
    ["\u65F6\u95F4\u672A\u77E5", "time unknown"],
    [" \u603B\u2264", " total \u2264"],
    ["\u53EF\u64A4\u9500", "revertible"],
    [" \xB7 \u5907\u6CE8 ", " \xB7 note "],
    [" \xB7 \u57DF ", " \xB7 domain "],
    ["\u5171 ", "total "],
    [" \u7247 \xB7 stale ", " shard(s) \xB7 stale "],
    [" \xB7 \u91CD\u5EFA\u9608\u503C ", " \xB7 rebuild threshold "],
    ["\u2014\u2014\u7248\u672C\u51B2\u7A81\uFF1A\u8BF7\u5237\u65B0\u9762\u677F\u5BF9\u9F50\u6700\u65B0\u7248\u672C\u540E\u91CD\u8BD5\uFF08\u539F\u7A3F\u4E0D\u8986\u76D6\uFF09", " \u2014 optimistic-lock conflict: reopen the panel, align revision, retry (draft not overwritten, never silently rebased)"],
    ["\uFF1B\u8D85\u9650 tag \u6536\u655B ", "; over-limit tags converged "],
    [" \u4E2A\uFF1A", ": "],
    ["\u64A4\u9500\u6700\u8FD1\u4E00\u6B21\u4E0A\u4E0B\u6587\u5206\u53C9\uFF08\u6062\u590D\u670D\u52A1\u7AEF\u5199\u524D\u5FEB\u7167\uFF09\uFF1F", "Revert the last context fork (restore the server-side pre-write snapshot)?"],
    ["\u2014\u2014contextidx \u672A\u88C5\u914D\uFF08\u9632\u5FA1\u9762\uFF09", " \u2014 contextidx not assembled (defensive face)"],
    ["\u9A8C\u8BC1\u4E2D", "Verifying"],
    ["\u6392\u961F\u4E2D", "Queued"],
    ["\u5F85\u542F\u52A8", "Pending start"],
    ["\u53D7\u963B", "Blocked"],
    ["\u5931\u8D25", "Failed"],
    ["\u5DF2\u5B8C\u6210", "Done"],
    ["\u5DF2\u53D6\u6D88", "Canceled"],
    ["\u672A\u547D\u540D\u4EFB\u52A1", "Unnamed task"],
    ["\u4F9D\u8D56 ", "Deps "],
    ["\u8D1F\u8D23 ", "Owner "],
    ["\u6700\u8FD1\u52A8\u4F5C\uFF1A\u2014", "Last action: \u2014"],
    ["\u5173\u8054\u4F1A\u8BDD", "Related session"],
    [" \u9879", " item(s)"],
    ["\u2190 \u4EFB\u52A1\u4E2D\u5FC3", "\u2190 Task center"],
    ["\u4EFB\u52A1\u4E0D\u5728\u518C", "Task not registered"],
    ["\u6682\u65E0\u6B64\u4EFB\u52A1\uFF08\u53EF\u80FD\u5DF2\u5B8C\u6210\u5F52\u6863\u6216 id \u6709\u8BEF\uFF09\u3002", "No such id in the snapshot task list (completed out of window, or wrong id)."],
    ["\u6700\u8FD1\u52A8\u4F5C", "Last action"],
    ["\u6700\u8FD1\u9519\u8BEF", "Last error"],
    ["\u4E25\u91CD", "Critical"],
    ["\u91CD\u8981", "Major"],
    ["\u4E00\u822C", "Minor"],
    ["\u672A\u5206\u7EA7", "Ungraded"],
    ["\u8BC1\u636E\u2014\u2014\u70B9\u51FB\u8FDB\u5165", "Evidence \u2014 click through"],
    ["\u8BC1\u636E", "Evidence"],
    ["\u672A\u547D\u540D\u5173\u6CE8\u9879", "Unnamed attention item"],
    ["\u667A\u80FD\u5F52\u7EB3", "Smart digest"],
    ["\u57FA\u7840\u6C47\u603B", "Basic summary"],
    ["\u751F\u6210\u65B9\u5F0F\u672A\u6CE8\u660E", "Generation method unstated"],
    [" \xB7 \u68C0\u67E5\u8303\u56F4 ", " \xB7 check scope "],
    [" \xB7 \u622A\u81F3 ", " \xB7 as of "],
    [" \xB7 \u7F3A\u5E2D\u6570\u636E\uFF1A", " \xB7 absent data: "],
    ["severity \u6392\u5E8F \xB7 \u8BC1\u636E\u53EF\u70B9 \xB7 \u6570\u636E\u53EF\u4FE1\u5EA6\u663E\u5F0F", "severity-sorted \xB7 clickable evidence \xB7 data trust stated"],
    ["\u672C\u7A97\u53E3\u65E0\u5173\u6CE8\u9879", "No attention items in this window"],
    ["\u5F53\u524D\u6CA1\u6709\u9700\u8981\u5173\u6CE8\u7684\u98CE\u9669\u9879\u3002", "attention aggregate zero-hit (explicit empty \u2014 never fakes risk)."],
    ["\u5F53\u524D\u5FAA\u73AF", "Current loop"],
    ["\u89E6\u53D1\u4E8B\u4EF6", "Fire event"],
    ["\u5F53\u524D\u65E0\u8FDB\u884C\u4E2D\u5FAA\u73AF", "No loop in progress"],
    ["\u5168\u90E8", "All"],
    ["\u7A3F", "Draft"],
    ["\u5347\u7EA7", "Escalate"],
    ["\u65E0\u4F1A\u8BDD", "No session"],
    ["\u4E8B\u4EF6", "Events"],
    ["\u7A3F\u222A\u5F85\u5BA1\u4E8B\u4EF6", "drafts \u222A pending events"],
    ["\u6253\u5F00\u5BF9\u8BDD\u7EBF\u7A0B", "Open dialog thread"],
    ["\u6B63\u5728\u8DD1\u7684 Workflows", "Running workflows"],
    ["\u6700\u8FD1\u9700\u8981\u7684\u5BA1\u6279", "Approvals needed recently"],
    ["\u6838\u5FC3 Agent", "Core agents"],
    ["\u7CFB\u7EDF\u89C6\u89D2\u5927\u5C4F\u2014\u2014\u6B63\u5728\u8DD1\u7684\u4EFB\u52A1\u3001\u6700\u8FD1\u9700\u8981\u7684\u5BA1\u6279\u3001\u6838\u5FC3 Agent \u4E00\u773C\u53EF\u8BFB", "System-view dashboard \u2014 running tasks, approvals needed, core agents at a glance"],
    ["Service", "Service"],
    ["Service \u89C6\u89D2\u5927\u5C4F\uFF08\u4E0E\u7CFB\u7EDF\u5927\u5C4F\u540C\u6784\u5E03\u5C40\u2014\u2014\u540E\u7EED ServiceAgent \u767B\u5F55\u5373\u843D\u6B64\u89C6\u56FE\uFF09", "Service-view dashboard (same layout as the system one \u2014 a later ServiceAgent login lands here)"],
    ["ServiceDesc \u5728\u518C\u2014\u2014agents=", "ServiceDesc registered \u2014 agents="],
    ["ServiceDesc \u5BF9\u8C61\u4E0D\u5728\u518C\uFF08\u8BCD\u5F62\u6295\u5F71\u81EA\u5FEB\u7167\u2014\u2014\u4E0D\u5192\u5145\u5728\u518C\uFF09", "ServiceDesc object not registered (projected from snapshot \u2014 never fakes registration)"],
    ["\u8FDB\u884C\u4E2D", "Running"],
    ["\u5355\u8282\u70B9\uFF08\u58F0\u660E\u672A\u6295\u5F71\u76F8\u4F4D\uFF09", "Single node (declared phases not projected)"],
    ["\u70B9\u51FB\u8282\u70B9\u67E5\u770B\u8BE5\u9636\u6BB5\u7684\u4E8B\u4EF6\u4E0E\u6D89\u53CA\u6587\u6863\u3002", "Click a node for that phase\u2019s event chain and involved docs; progress refreshes live via patches."],
    ["\u7A97\u53E3\u5185\u65E0\u8BE5\u76F8\u4F4D\u4E8B\u4EF6\u3002", "No events for this phase in window (honest empty)."],
    ["\u6D89\u53CA", "Involved"],
    ["\u4E8B\u4EF6\u94FE\uFF08\u5F52\u6863\u7A97\u53E3\uFF09\uFF1A", "Event chain (archive window):"],
    ["\u672A\u77E5", "Unknown"],
    [" \u6B21\u8FD0\u884C", " run(s)"],
    ["\u805A\u5408\u9762\uFF08\u5FEB\u7167 workflows/runs\uFF09\uFF1B\u8FD0\u884C\u72EC\u7ACB\u9875\u89C1\u8BE6\u60C5", "Aggregate face (snapshot workflows/runs); runs have their own detail pages"],
    ["\u6682\u65E0\u5728\u518C\u5DE5\u4F5C\u6D41", "No registered workflows"],
    ["\u5FEB\u7167 workflows \u805A\u5408\u4E3A\u7A7A\uFF08\u4E0D\u88C5\u5047\uFF09\u3002", "Snapshot workflows aggregate is empty (no faking)."],
    ["\u5DE5\u4F5C\u6D41\u8BE6\u60C5", "Workflow detail"],
    ["\u2190 \u5DE5\u4F5C\u6D41", "\u2190 Workflows"],
    ["\u5DE5\u4F5C\u6D41\u4E0D\u5728\u518C", "Workflow not registered"],
    ["\u6682\u65E0\u6B64\u5DE5\u4F5C\u6D41\uFF08\u53EF\u80FD\u5DF2\u4E0B\u7EBF\u6216\u540D\u79F0\u6709\u8BEF\uFF09\u3002", "No such workflow in the snapshot aggregate (offline, or wrong name)."],
    ["\u76EE\u6807", "Target"],
    ["\u8FD0\u884C\u8BB0\u5F55\u4E0D\u5728\u518C", "Run not registered"],
    ["\u6682\u65E0\u6B64\u8FD0\u884C\u8BB0\u5F55\uFF08\u53EF\u80FD\u5DF2\u5F52\u6863\uFF09\u3002", "No such run in the snapshot runs aggregate (out of window)."],
    ["\u7B49\u5BA1\u6279", "Awaiting approval"],
    ["\u8FD0\u884C\u4E2D", "Running"],
    ["\u53BB\u5BA1\u6279 \u2192", "To approvals \u2192"],
    ["\u5F53\u524D\u76F8\u4F4D", "Current phase"],
    ["\u4E8B\u4EF6\u65F6\u95F4\u7EBF", "Event timeline"],
    ["\u5355\u8282\u70B9", "Single node"],
    ["\u6B63\u5728\u53D1\u751F\u2014\u2014\u63A8\u8FDB\u7531 patch \u5B9E\u65F6\u5237\u65B0", "Happening now \u2014 progress refreshes live via patches"],
    ["\u5386\u53F2\u76F8\u4F4D\u2014\u2014\u4E8B\u4EF6\u6309\u5F52\u6863\u7A97\u53E3\u56DE\u653E", "Past phase \u2014 events replayed from the archive window"],
    ["\u53BB\u5904\u7406\u8FD9\u6761\u5BA1\u6279", "Handle this approval"],
    ["\u6682\u65E0\u76F8\u5173\u4E8B\u4EF6\u2014\u2014\u6D41\u7A0B\u63A8\u8FDB\u65F6\u6B64\u5904\u5B9E\u65F6\u66F4\u65B0\u3002", "No events in the archive window yet \u2014 rows land here live as the workflow advances."],
    /* ── 会话屏 v4（C 批）EN 回填（chat.js 新词形——C-11 双语随行） ── */
    ["\u6709\u4EC0\u4E48\u53EF\u4EE5\u5E2E\u5FD9\u7684\uFF1F", "How can I help?"],
    ["\u627E\u6587\u6863\uFF0F\u66F4\u65B0\u6587\u6863\uFF0F\u63D0\u4EA4\u9700\u6C42\u2014\u2014\u4F1A\u8BDD\u7531\u5BBF\u4E3B\u5EFA\u6863\uFF0C\u8F6C\u5F55\u5373\u771F\u5B9E journal\u3002", "Find docs / update docs / submit requests \u2014 sessions are created by the host; the transcript is the real journal."],
    ["\u8FDB\u884C\u4EE3\u7801\u5BA1\u6838\uFF0C\u770B\u770B\u5B9E\u9645\u548C\u6587\u6863\u5DEE\u8DDD\u6709\u591A\u5927\uFF0C\u5E76\u8FDB\u884C\u4FEE\u590D", "Run a code audit, measure the actual-vs-doc gap, and fix it"],
    ["\u627E\u51FA\u6240\u6709 drift \u7684\u6587\u6863\u5E76\u751F\u6210\u5BF9\u8D26\u6E05\u5355", "Find all drifted docs and produce a reconciliation list"],
    ["\u628A\u5F53\u524D\u5FAA\u73AF\u7684\u72B6\u6001\u5199\u6210\u4E00\u4EFD SessionLog \u6458\u8981", "Summarize the current loop into a SessionLog"],
    ["\u68C0\u67E5 workflow \u5F15\u64CE\u7684\u8FD0\u884C\u8BB0\u5F55\uFF0C\u6807\u51FA\u5931\u8D25\u539F\u56E0", "Inspect workflow engine runs and flag failure causes"],
    ["\u63D0\u4EA4\u9700\u6C42\uFF0F\u627E\u6587\u6863\uFF0F\u66F4\u65B0\u6587\u6863\u2026\u2026", "Submit a request / find a doc / update a doc\u2026"],
    ["\u53D1\u9001", "Send"],
    ["\u505C\u6B62", "Stop"],
    ["Enter \u53D1\u9001 \xB7 Shift+Enter \u6362\u884C", "Enter to send \xB7 Shift+Enter for newline"],
    ["\u6B63\u5728\u8C03\u7528", "Calling"],
    ["\u5DF2 {n} \u79D2", "({n}s)"],
    ["agent \u8FD0\u884C\u4E2D\u2026", "agent working\u2026"],
    ["\u6392\u961F {n}", "queue {n}"],
    ["\u8FFD\u52A0 {n}", "+{n} sent"],
    ["Provider", "Provider"],
    ["\u914D\u7F6E OpenAIProvider\uFF08api-key \u5F15\u7528 / base \u5730\u5740 / \u6A21\u578B\u5206\u7EA7\uFF09", "Configure OpenAIProvider (api-key ref / base URL / model tiers)"],
    ["\u4E09\u6BB5\u5F0F Session \u73B0\u8BFB\u4E2D\u2026", "Loading three-stage session\u2026"],
    ["\u4E0A\u4E0B\u6587\u73B0\u8BFB\u4E2D\u2026\uFF08\u9009\u5B9A\u4F1A\u8BDD\u540E\u88C5\u914D\uFF09", "Loading context\u2026 (assembled once a session is selected)"],
    ["\u4E09\u6BB5\u5F0F Session", "Three-stage Session"],
    ["\u5DF2\u6301\u4E45\u5316", "persisted"],
    ["\u7F3A\u7701\uFF08\u672A\u6301\u4E45\u5316\uFF09", "defaults (not persisted)"],
    ["\u7B2C\u4E00\u6BB5\xB7\u80CC\u666F\u88C5\u914D", "Stage 1 \xB7 Context assembly"],
    ["\u7B2C\u4E8C\u6BB5\xB7\u9884\u66F4\u65B0\u6267\u884C", "Stage 2 \xB7 Pre-update execution"],
    ["\u7B2C\u4E09\u6BB5\xB7\u5C5E\u6027\u56DE\u6574", "Stage 3 \xB7 Attribute writeback"],
    ["\u5B8C\u6210", "done"],
    ["\u8FDB\u884C\u4E2D", "running"],
    ["\u5F85\u89E6\u53D1", "pending"],
    ["\u4E0A\u4E0B\u6587\u9762\u5DF2\u73B0\u8BFB\uFF08\u7F13\u5B58 chunks/\u5206\u7247\u89C1\u53F3\u680F\uFF09", "Context face loaded (cached chunks/shards in the right rail)"],
    ["\u7B49\u5F85\u4F1A\u8BDD\u4E0A\u4E0B\u6587\u9762\u88C5\u914D", "waiting for session context assembly"],
    ["\u4F1A\u8BDD\u8FD0\u884C\u4E2D\uFF08\u53CD\u590D ask \u8BE2\u95EE\uFF1B\u4EA7\u51FA=\u6587\u6863\u9884\u66F4\u65B0 draft\uFF09", "session running (ask loop; output = document pre-update drafts)"],
    ["\u9884\u66F4\u65B0\u5DF2 apply\u2014\u2014\u8FDB\u5165\u56DE\u6574", "pre-update applied \u2014 entering writeback"],
    ["\u7B49\u5F85\u8F93\u5165\u6216\u7B49\u5F85 apply\uFF08\u4F1A\u8BDD\u6CA1\u6709\u547D\u4EE4\u2014\u2014\u53EA\u4EA7 draft\uFF09", "waiting for input or apply (no commands in session \u2014 drafts only)"],
    ["apply \u540E\u81EA\u52A8\u89E6\u53D1\uFF08\u9ED8\u8BA4\u56DE\u6574\uFF1Aref/tag/\u91CD\u8981\u7A0B\u5EA6/metainfo\uFF09", "auto-triggered after apply (writeback: ref/tag/importance/metainfo)"],
    [" \xB7 importance +", " \xB7 importance +"],
    [" \xB7 tag+", " \xB7 tag+"],
    ["\u81EA\u52A8\u56DE\u6574", "auto writeback"],
    ["\u56DE\u6574\u8D26\uFF08\u7B2C\u4E09\u6BB5\uFF09", "Writeback log (stage 3)"],
    ["\u4E09\u6BB5\u63A7\u5236\uFF08\u63CF\u8FF0/\u9884\u7B97/\u5F00\u5173\u2014\u2014PUT /v1/session-config\uFF09", "Stage control (descriptions/budget/toggles \u2014 PUT /v1/session-config)"],
    ["\u4E09\u6BB5\u63CF\u8FF0\u4E0E\u5F00\u5173=\u4E09\u6BB5\u5F0F\u884C\u4E3A\u7684\u627F\u8F7D\u9762\uFF08scheme runtime.context \u540C\u6E90\uFF1B\u6B64\u5904\u4E3A\u6570\u636E\u5E93\u627F\u8F7D\u884C\u2014\u2014\u4FDD\u5B58\u5373\u751F\u6548\uFF09\u3002", "Stage descriptions and toggles carry the three-stage behavior (same source as scheme runtime.context; this row is the DB carrier \u2014 effective on save)."],
    ["\u80CC\u666F\u9884\u7B97\uFF08\u5B57\uFF09", "Background budget (chars)"],
    ["\u56DE\u6574\u5C5E\u6027\u57DF", "Writeback attr scope"],
    ["\u4F1A\u8BDD Provider", "Session provider"],
    ["\u4F1A\u8BDD\u6A21\u578B", "Session model"],
    ["\u80CC\u666F\u9010\u5757\u6EAF\u6E90", "Per-block attribution"],
    ["\u7B2C\u4E09\u6BB5\u5C5E\u6027\u81EA\u52A8\u56DE\u6574", "Stage-3 auto attribute writeback"],
    ["\u4FDD\u5B58\u4E09\u6BB5\u63A7\u5236", "Save stage control"],
    ["\u4FDD\u5B58\u4E2D\u2026", "saving\u2026"],
    ["\u5DF2\u4FDD\u5B58\uFF08", "saved ("],
    ["\uFF09", ")"],
    ["\u4FDD\u5B58\u5931\u8D25\uFF1A", "save failed: "],
    ["\u80CC\u666F\u9884\u7B97 ", "Background budget "],
    [" \u5B57 \xB7 \u9010\u5757\u6EAF\u6E90 ", " chars \xB7 per-block attribution "],
    [" \xB7 \u5C5E\u6027\u56DE\u6574 ", " \xB7 attr writeback "],
    [" \xB7 Provider ", " \xB7 Provider "],
    ["\u5F00", "on"],
    ["\u5173", "off"],
    ["\u5DF2\u6DF1\u5EA6\u601D\u8003", "Deep thinking"],
    ["\u5DF2\u8C03\u7528", "Called"],
    ["\u5B8C\u6210", "done"],
    ["\u6267\u884C\u4E2D\u2026", "running\u2026"],
    ["\u2500\u2500\u2500\u2500 \u7ED3\u679C \u2500\u2500\u2500\u2500", "\u2500\u2500\u2500\u2500 result \u2500\u2500\u2500\u2500"],
    ["\u9001\u8FBE\u4E2D\u2026", "Delivering\u2026"],
    ["\u9001\u8FBE\u5931\u8D25\uFF08", "Delivery failed ("],
    ["\u91CD\u8BD5", "Retry"],
    ["\u6CE8\u5165\u4E0A\u4E0B\u6587", "Injected context"],
    ["\u505C\u6B62\u901A\u8DEF\u672A\u6CE8\u5165\uFF08chatfaces.js.stop \u7F3A\u5E2D\uFF09\u2014\u2014\u663E\u5F0F\u7F3A\u5E2D\uFF0C\u4E0D\u5192\u5145\u3002", "Stop channel not injected (chatfaces.js.stop absent) \u2014 explicitly absent, never faked."],
    ["\u5DF2\u8BF7\u6C42\u505C\u6B62\u2014\u2014\u4F1A\u8BDD\u56DE\u5230\u7A7A\u95F2\u540E\u8F6C\u5F55\u53EF\u89C1\u4E2D\u65AD\u8BCD\u5F62\u3002", "Stop requested \u2014 the interrupt word form appears in the transcript once the session goes idle."],
    ["\u505C\u6B62\u5931\u8D25\uFF08", "Stop failed ("],
    ["\u65B0\u4F1A\u8BDD", "New session"],
    ["\u65B0\u4F1A\u8BDD\u5931\u8D25\uFF08", "New session failed ("],
    ["\u4F1A\u8BDD\u6E05\u5355\u4E3A\u7A7A\u2014\u2014\u70B9\u300C\u65B0\u4F1A\u8BDD\u300D\u7531\u5BBF\u4E3B\u5EFA\u6863\u3002", "No sessions yet \u2014 click New session and the host creates one."],
    ["\u4F1A\u8BDD\u6E05\u5355\u6682\u4E0D\u53EF\u8BFB\uFF08", "Session list unreadable ("],
    ["\u4F1A\u8BDD\u73B0\u8BFB\u4E2D\u2026", "Loading sessions\u2026"],
    ["\u4ECA\u5929", "Today"],
    ["\u66F4\u65E9", "Earlier"],
    ["\u7A7A\u767D", "blank"],
    ["\u7A7A\u95F2", "Idle"],
    ["\u8FD0\u884C\u4E2D", "Running"],
    ["\u8F6C\u5F55\u73B0\u8BFB\u4E2D\u2026", "Loading transcript\u2026"],
    ["\u8F6C\u5F55\u9762\u6682\u4E0D\u53EF\u8BFB\uFF08", "Transcript face unreadable ("],
    ["\uFF09\u2014\u2014\u4E0D\u5192\u5145\u7A7A\u3002", ") \u2014 never fakes an empty state."],
    ["\uFF08\u7A7A\u56DE\u590D\uFF09", "(empty reply)"],
    ["\u5DE5\u5177", "Tool"],
    ["\u81EA\u6709\u5BF9\u8BDD\u9762\u2014\u2014\u771F\u5B9E\u8F6C\u5F55\u4E0E\u53D1\u9001\uFF1B\u4F1A\u8BDD id \u7531\u5BBF\u4E3B\u53D1\uFF0C\u6C38\u4E0D\u624B\u586B", "Own chat face \u2014 real transcript and delivery; session ids come from the host, never hand-typed"],
    ["\u5BF9\u8BDD\u901A\u8DEF\u672A\u6CE8\u5165\uFF08chatfaces.js \u7F3A\u5E2D\uFF09\u2014\u2014\u8F6C\u5F55\u4E0E\u53D1\u9001\u9762\u663E\u5F0F\u7F3A\u5E2D\uFF0C\u4E0D\u5192\u5145\u3002", "Dialog channel not injected (chatfaces.js absent) \u2014 transcript & delivery faces explicitly absent, never faked."],
    ["\u5F00\u53D1\u8005", "Developer"],
    ["\u2014\u2014\u4EFB\u52A1\u8BE6\u60C5", " \u2014 task detail"],
    ["\u5FEB\u7167\u4E09\u6E90\uFF08Task.agentSession \u222A \u4E8B\u4EF6 agentSession\uFF09\u53CD\u67E5\u65E0\u5728\u518C\u5173\u8054\u2014\u2014\u4F1A\u8BDD\u578B\u72EC\u7ACB\u4F1A\u8BDD\uFF08\u96F6\u7F16\u9020\uFF09\u3002", "Reverse lookup across the three snapshot sources (Task.agentSession \u222A event agentSession) finds no registered link \u2014 a standalone session (zero fabrication)."],
    ["\u5173\u8054\uFF1A", "Linked: "],
    ["root \u4F1A\u8BDD\uFF08\u9ED8\u8BA4\uFF09", "root session (default)"],
    ["Enter \u53D1\u9001 \xB7 \u6D88\u606F\u9001\u8FBE\u7CFB\u7EDF\u4F1A\u8BDD\uFF1B\u5904\u7406\u7ED3\u8BBA\u4EE5\u6587\u6863\u4E0E\u6D41\u7A0B\u72B6\u6001\u5448\u73B0", "Enter to send \xB7 queue semantics (delivered to the root agent session; the conclusion lands in docs by the agent, visible via status badge and node graph)"],
    ["\u5DF2\u9001\u8FBE\uFF08queue\uFF09\u2014\u2014root agent \u4F1A\u8BDD\u5DF2\u53D7\u7406\uFF1B\u7ED3\u8BBA\u7ECF\u72B6\u6001\u5FBD\u4E0E\u8282\u70B9\u56FE\u53EF\u89C1\u3002", "Delivered (queue) \u2014 root agent session accepted; watch the status badge and node graph for the conclusion."],
    ["\uFF09", ")"],
    /* ── 2026-11-27 首页=对话（DeepSeek chat 模式）── */
    ["\u5BF9\u8BDD", "Chat"],
    ["\u5F00\u59CB", "Start"],
    ["\u603B\u89C8", "Overview"],
    ["\u56DE\u9996\u9875", "Back to home"],
    ["\u76F4\u63A5\u8BF4\u9700\u6C42\u2014\u2014\u627E\u6587\u6863\u3001\u66F4\u65B0\u6587\u6863\u3001\u63D0\u4EA4\u53D8\u66F4\uFF1B\u7ED3\u8BBA\u843D\u5230\u6587\u6863\u4E0E\u6D41\u7A0B\u3002", "Just ask \u2014 find docs, update docs, submit changes; results land in docs and workflows."],
    ["\u63D0\u4EA4\u9700\u6C42\u3001\u627E\u6587\u6863\u3001\u66F4\u65B0\u6587\u6863\u2026\u2026", "Ask for a change, find docs, update docs\u2026"],
    ["\u5BF9\u8BDD\u8F93\u5165", "Chat input"],
    ["\u627E\u6587\u6863", "Find docs"],
    ["\u66F4\u65B0\u6587\u6863", "Update docs"],
    ["\u56FE\u7247", "Image"],
    ["\u9644\u52A0\u56FE\u7247\uFF08\u22648MB\uFF09", "Attach image (\u22648MB)"],
    ["\u6B63\u5728\u521B\u5EFA\u4F1A\u8BDD\u2026", "Creating chat\u2026"],
    ["\u65B0\u4F1A\u8BDD\u521B\u5EFA\u5931\u8D25", "Failed to create chat"],
    ["\u6B63\u5728\u53D1\u9001\u2026", "Sending\u2026"],
    ["\u5DF2\u9001\u8FBE\uFF0C\u6B63\u5728\u6253\u5F00\u4F1A\u8BDD\u2026", "Delivered \u2014 opening chat\u2026"],
    ["\u53D1\u9001\u5931\u8D25", "Send failed"],
    ["\u56FE\u7247\u4E0A\u4F20\u901A\u8DEF\u672A\u6CE8\u5165", "Image upload path not injected"],
    ["\u56FE\u7247\u4E0A\u4F20\u4E2D\u2026", "Uploading image\u2026"],
    ["\u56FE\u7247\u5DF2\u9644\u52A0\uFF08\u968F\u6D88\u606F\u53D1\u9001\uFF09", "Image attached (sent with your message)"],
    ["\u56FE\u7247\u4E0A\u4F20\u5931\u8D25", "Image upload failed"],
    ["\u56FE\u7247\u8D85\u8FC7 8MB \u4E0A\u9650", "Image exceeds the 8MB limit"],
    ["\u6700\u8FD1\u4F1A\u8BDD", "Recent chats"],
    ["\u5168\u90E8\u4F1A\u8BDD", "All chats"],
    /* ── 2026-11-27 Workflow 详情（真实定义/节点/任务/会话）── */
    ["\u4EFB\u52A1", "Tasks"],
    ["\u4EFB\u52A1\u951A\uFF08\u975E sessionId\uFF09", "Task anchor (not a sessionId)"],
    ["\u4F1A\u8BDD\u672A\u8BB0\u5F55", "No session recorded"],
    ["\u542F\u7528", "Enabled"],
    ["\u505C\u7528", "Disabled"],
    ["\u5230\u8FBE", "Arrivals"],
    ["\u5230\u8FC7", "Visits"],
    ["\u5F53\u524D", "Current"],
    ["\u5DF2\u8D70\u8FC7", "Completed"],
    ["\u672A\u5230\u8FBE", "Not reached"],
    ["\u6700\u65B0", "Latest"],
    ["\u6700\u8FD1", "Recent"],
    ["\u6700\u8FD1\u8FD0\u884C", "Latest run"],
    ["\u6765\u6E90", "Source"],
    ["\u5B9A\u4E49", "Definition"],
    ["\u5B9A\u4E49/\u8FD0\u884C/\u4EFB\u52A1\u5747\u4E3A\u73B0\u8BFB\u771F\u5B9E\u6570\u636E", "Definition, runs and tasks are live data"],
    ["\u5728\u518C", "Registered"],
    ["\u7F3A\u5E2D", "Absent"],
    ["\u6267\u884C\u9762", "Execution"],
    ["\u6267\u884C\u8282\u70B9\uFF08ConnectorOps \u9AA8\u67B6\uFF09", "Execution nodes (ConnectorOps skeleton)"],
    ["\u8282\u70B9\uFF08\u76F8\u4F4D\u94FE\uFF09", "Nodes (phase chain)"],
    ["\u8282\u70B9\u4F1A\u8BDD", "Node sessions"],
    ["\u8282\u70B9\u4F1A\u8BDD\u73B0\u8BFB\u4E2D\u2026", "Loading node sessions\u2026"],
    ["\u8282\u70B9\u4F1A\u8BDD\u8BFB\u53D6\u5931\u8D25\u2014\u2014\u4E0D\u5192\u5145\u5DF2\u8BB0\u5F55\u3002", "Failed to read node sessions \u2014 not shown as recorded."],
    ["\u8BE5\u8FD0\u884C\u672A\u5173\u8054 Task \u5BF9\u8C61\u2014\u2014\u8282\u70B9\u4F1A\u8BDD\u672A\u8BB0\u5F55\u3002", "This run has no linked Task object \u2014 no node session recorded."],
    ["\u8FD0\u884C\u8BB0\u5F55", "Runs"],
    ["\u8FD0\u884C\u672A\u5173\u8054 Task \u5BF9\u8C61\uFF08\u89E6\u53D1\u6E90=Proposal/Template/\u5176\u4ED6\u2014\u2014\u89C1\u8FD0\u884C involved\uFF09\u3002", "Runs are not linked to Task objects (trigger source is Proposal/Template/other \u2014 see run involved)."],
    ["\u8BE5 workflow \u6682\u65E0\u8FD0\u884C\u8BB0\u5F55\u2014\u2014", "No runs for this workflow \u2014 "],
    ["\u6267\u884C\u9762\u672A\u7ED1\u5B9A\uFF08ConnectorOps \u9AA8\u67B6\u7F3A\u5E2D\uFF09\uFF0C\u4EFB\u52A1\u4E0D\u4F1A\u62C9\u8D77 run\u3002", "Execution is unbound (ConnectorOps skeleton absent); tasks do not start runs."],
    ["\u6267\u884C\u9762\u672A\u7ED1\u5B9A\uFF08ConnectorOps \u9AA8\u67B6\u7F3A\u5E2D\uFF09\u2014\u2014workflow apply \u540E\u7684\u843D\u5730 agent \u5F85\u63A5\u7EBF\uFF1B\u5F53\u524D\u8282\u70B9=\u76F8\u4F4D\u94FE\uFF0C\u8282\u70B9\u7EA7\u4F1A\u8BDD\u672A\u5355\u72EC\u8BB0\u5F55\u3002", "Execution is unbound (ConnectorOps skeleton absent) \u2014 the workflow apply agent is pending; current nodes are the phase chain, with no separate node sessions recorded."],
    ["\u8282\u70B9\u4F1A\u8BDD\u951A=history.sessionId \u2192 run.sessionId \u2192 Task journal \u53CD\u67E5\uFF08\u7A7A=\u672A\u8BB0\u5F55\uFF0C\u4E0D\u7F16\u9020\uFF1B\u65E7 task.agentSession=taskId \u4E0D\u5F53 session\uFF09\u3002", "Node session anchor: history.sessionId \u2192 run.sessionId \u2192 Task journal lookup (empty = not recorded; legacy task.agentSession=taskId is not treated as a session)."],
    ["\u5DE5\u4F5C\u6D41\u73B0\u8BFB\u4E2D\u2026\uFF08\u5B9A\u4E49/\u8FD0\u884C/\u4EFB\u52A1\uFF09", "Loading workflow\u2026 (definition/runs/tasks)"],
    ["\u5DE5\u4F5C\u6D41\u73B0\u8BFB\u5931\u8D25\uFF1A", "Failed to load workflow: "],
    ["\u5DE5\u4F5C\u6D41\u9762\u6682\u4E0D\u53EF\u8BFB", "Workflow data unavailable"],
    ["GET /v1/workflows \u4E0E /v1/workflows/runs \u5747\u672A\u8FD4\u56DE\u6570\u636E\u2014\u2014", "Neither GET /v1/workflows nor /v1/workflows/runs returned data \u2014 "],
    ["\u76F8\u8FD1\u5DE5\u4F5C\u6D41", "Similar workflows"],
    ["\u8BCD\u5F62\u517C\u5BB9", "Alias"],
    ["\u7F16\u8F91 Workflow", "Edit Workflow"],
    ["\u7F16\u8F91 Workflow \u58F0\u660E", "Edit Workflow declaration"],
    ["\u7F16\u8F91\u5B9A\u4E49\uFF08workflow.md\uFF09", "Edit definition (workflow.md)"],
    ["\u76F8\u4F4D\u8868\uFF08scheme\uFF09", "Phase table (scheme)"],
    ["\u8BBE\u8BA1\u6587\u6863\u5DE5\u4F5C\u53F0", "Design doc workspace"],
    ["\u70B9\u51FB\u8282\u70B9\uFF1A\u5230\u8FBE\u8BE5\u8282\u70B9\u7684\u8FD0\u884C/\u4EFB\u52A1\u4E0E\u771F\u5B9E\u4F1A\u8BDD\u951A\u3002", "Click a node to see the runs/tasks and real session anchors that reached it."],
    ["\u6682\u65E0\u8FD0\u884C\u5230\u8FBE\u8BE5\u8282\u70B9\u3002", "No runs reached this node."],
    ["\u8BE5 workflow \u672A\u6295\u5F71\u76F8\u4F4D\u94FE\u3002", "This workflow has no projected phase chain."],
    ["\u73AF\u5185 ", "In ring "],
    ["runs", "runs"],
    ["ConnectorOps", "ConnectorOps"],
    ["\u9AA8\u67B6", "Skeleton"],
    ["\u751F\u6548", "Effective"],
    ["\u8FD0\u884C\u4F1A\u8BDD", "Run session"],
    ["\u4EFB\u52A1\u4F1A\u8BDD", "Task session"],
    ["\u65E0\u4F1A\u8BDD\uFF08\u672A\u5230\u8FBE\uFF09", "No session (not reached)"],
    ["\u5171 {n} \u4E2A\u771F\u5B9E\u4F1A\u8BDD", "{n} real sessions"],
    // —— D 批（2026-11-27 DeepSeek harness 展示类型对账，chat-ds.md）——
    ["\u540E\u53F0\u4EFB\u52A1\u72B6\u6001\u66F4\u65B0", "Background job update"],
    ["\u63D2\u4EF6\u72B6\u6001\u66F4\u65B0", "Plugin update"],
    ["\u81EA\u52A8\u5316\u4EFB\u52A1", "Scheduled task"],
    ["\u5B50\u4EFB\u52A1\u72B6\u6001\u66F4\u65B0", "Subtask update"],
    ["\u5B50\u667A\u80FD\u4F53\u5DF2\u5B8C\u6210\u4EFB\u52A1", "Subagent finished the task"],
    ["\u5B50\u667A\u80FD\u4F53\u4EFB\u52A1\u5DF2\u505C\u6B62", "Subagent task stopped"],
    ["\u5B50\u667A\u80FD\u4F53\u6267\u884C\u5931\u8D25", "Subagent failed"],
    ["\u5B50\u667A\u80FD\u4F53\u5DF2\u8FBE token \u4E0A\u9650", "Subagent hit the token cap"],
    ["\u5B50\u667A\u80FD\u4F53\u62D2\u7EDD\u6267\u884C", "Subagent refused"],
    ["\u6536\u5230\u6267\u884C\u8BF7\u6C42", "Execution request received"],
    ["\u7EE7\u7EED\u6267\u884C\u76EE\u6807", "Continuing the goal"],
    ["\u6536\u5230\u4EFB\u52A1\u6D88\u606F", "Task message received"],
    ["\u6536\u5230\u56E2\u961F\u6D88\u606F", "Team message received"],
    ["\u6536\u5230 GitHub \u4E8B\u4EF6", "GitHub event received"],
    ["\u6536\u5230\u5916\u90E8\u4E8B\u4EF6", "External event received"],
    ["\u8FD9\u6761\u901A\u77E5\u89E6\u53D1\u4E86\u672C\u8F6E\u56DE\u590D\u3002", "This notice triggered this turn."],
    ["\u4E0A\u4E0B\u6587\u6CE8\u5165", "Context injection"],
    ["\u8DE8\u4F1A\u8BDD\u53EC\u56DE", "Cross-session recall"],
    ["\u5DF2\u5B8C\u6210\uFF0C\u7528\u65F6 ", "Completed in "],
    ["\u5904\u7406\u5931\u8D25", "Failed"],
    ["\u5DF2\u505C\u6B62", "Stopped"],
    ["\u5DF2\u8FBE\u5230\u8F93\u51FA token \u4E0A\u9650", "Output token cap reached"],
    ["\u56DE\u7B54\u88AB\u622A\u65AD\uFF0C\u5DF2\u6709\u8F93\u51FA\u4FDD\u7559\u5728\u5BF9\u8BDD\u4E2D\u3002\u53D1\u9001\u300C\u7EE7\u7EED\u300D\u53EF\u8BA9\u6A21\u578B\u63A5\u7740\u8F93\u51FA\u3002", "The reply was truncated; what was generated stays in the thread. Send \u201Ccontinue\u201D to let the model resume."],
    ["\u6DF1\u5EA6\u6C42\u7D22\u4E2D\u2026", "Deep thinking\u2026"],
    ["\u5C0F\u65F6", "h"],
    ["\u5206", "m"],
    ["\u79D2", "s"],
    ["{m}\u6708{d}\u65E5", "{m}/{d}"],
    ["{y}\u5E74{m}\u6708{d}\u65E5", "{y}/{m}/{d}"],
    ["\u8FD0\u884C\u547D\u4EE4", "Run command"],
    ["\u8BFB\u53D6", "Read"],
    ["\u8BFB\u53D6\u56FE\u7247", "Read image"],
    ["\u5199\u5165", "Write"],
    ["\u7F16\u8F91", "Edit"],
    ["\u641C\u7D22\u6587\u4EF6\u5185\u5BB9", "Search file contents"],
    ["\u67E5\u627E\u6587\u4EF6", "Find files"],
    ["\u7F51\u9875\u641C\u7D22", "Web search"],
    ["\u7F51\u9875\u83B7\u53D6", "Web fetch"],
    ["\u66F4\u65B0\u4EFB\u52A1\u6E05\u5355", "Update task list"],
    ["\u63D0\u95EE", "Ask"],
    ["\u5931\u8D25", "Failed"],
    ["\u9000\u51FA\u7801 {code}", "Exit code {code}"],
    ["\u4FE1\u53F7 {signal}", "Signal {signal}"],
    ["\u5931\u8D25\u539F\u56E0\uFF1A", "Failure: "],
    ["{n} \u884C", "{n} lines"],
    ["{n} \u4E2A\u6587\u4EF6\u53D8\u66F4", "{n} file changes"],
    ["{shown} \u5904\u5339\u914D \xB7 {files} \u4E2A\u6587\u4EF6", "{shown} matches \xB7 {files} files"],
    ["{shown} \u4E2A\u8DEF\u5F84", "{shown} paths"],
    ["\u65E0\u7ED3\u679C", "No results"],
    ["\u7528\u91CF {total}", "Usage {total}"],
    ["\u672C\u8F6E\u7528\u91CF", "Turn usage"],
    ["\u672A\u7F13\u5B58\u8F93\u5165", "Uncached input"],
    ["\u7F13\u5B58\u8BFB\u53D6", "Cache read"],
    ["\u7F13\u5B58\u5199\u5165", "Cache write"],
    ["\u8F93\u51FA", "Output"],
    ["\uFF08\u5176\u4E2D\u63A8\u7406 {tokens}\uFF09", "(reasoning {tokens})"],
    ["\u601D\u8003", "Thought"],
    ["{done}/{total} \u5DF2\u5B8C\u6210", "{done}/{total} done"],
    ["\u6B63\u5728\u91CD\u8BD5\u6A21\u578B\u8BF7\u6C42", "Retrying model request"],
    ["\u5DF2\u8BF7\u6C42\u7EE7\u7EED\u751F\u6210", "Continuation requested"],
    ["{label}\uFF08{retry}/{maximum}\uFF09", "{label} ({retry}/{maximum})"],
    ["\u590D\u5236", "Copy"],
    ["\u590D\u5236\u6210\u529F", "Copied"],
    ["\u4EE3\u7801\u5757", "Code"],
    ["\u81EA\u52A8\u6362\u884C", "Wrap"],
    ["\u53D6\u6D88\u81EA\u52A8\u6362\u884C", "Unwrap"],
    ["\u5F15\u7528", "Cited"],
    ["\u4E0A\u4E0B\u6587", "Context"],
    ["\u6253\u5F00", "Open"],
    ["\u5173\u95ED", "Close"],
    ["\u5728\u6587\u6863\u5E93\u4E2D\u6253\u5F00", "Open in doc library"],
    ["\u73B0\u8BFB\u4E2D\u2026", "Reading\u2026"],
    ["\u73B0\u8BFB\u5931\u8D25\uFF08", "Read failed ("],
    ["\u672C\u6587\u6863\u6682\u65E0\u6B63\u6587\u3002", "This document has no body yet."],
    ["\u672A\u5728\u6587\u6863\u5E93\u547D\u4E2D\uFF08design-docs \u767D\u540D\u5355\u4ED3\u6811\u65E0\u6B64\u8DEF\u5F84\uFF09\u2014\u2014\u6765\u6E90\u6458\u5F55\u89C1\u5BF9\u8BDD\u539F\u6587\u3002", "Not found in the doc library (no such path in the design-docs whitelist tree) \u2014 see the source excerpt in the thread."],
    ["\u5BF9\u8C61\u9762\u6682\u4E0D\u53EF\u8BFB\uFF08", "Object face unreadable ("],
    ["\u6587\u6863\u76F4\u8BFB\u5931\u8D25\uFF08", "Doc read failed ("],
    ["\u4EFB\u52A1\u8BE6\u60C5\xB7\u6700\u8FD1\u52A8\u4F5C\u7ED3\u6784\u5316\u5361\uFF082026-11-28\uFF09", "Task detail structured action card (2026-11-28)"],
    ["\u7ED3\u679C \xB7 \u4F1A\u8BDD\u73B0\u52BF \xB7 \u751F\u547D\u5468\u671F\u65F6\u95F4\u7EBF", "Result \xB7 Session liveness \xB7 Lifecycle timeline"],
    ["\u7ED3\u679C\u6CE8\u8BB0", "Result note"],
    ["\u5173\u8054\u4F1A\u8BDD\u73B0\u52BF", "Session liveness"],
    ["\u4E0D\u5728\u5FC3\u8DF3\u5728\u518C", "Not in heartbeat ledger"],
    ["\u6700\u8FD1\u5FC3\u8DF3 ", "Last beat "],
    ["\u751F\u547D\u5468\u671F\uFF08\u4E8B\u4EF6\u5F52\u6863\u73B0\u8BFB\uFF09", "Lifecycle (read live from event archive)"],
    ["\u65F6\u95F4\u7EBF\u62C9\u53D6\u5931\u8D25\uFF1A", "Timeline fetch failed: "],
    ["\u5F52\u6863\u7A97\u5185\u6682\u65E0\u672C\u4EFB\u52A1\u4E8B\u4EF6\u3002", "No events for this task in the archive window."],
    ["\u672A\u547D\u540D\u4E8B\u4EF6", "Unnamed event"],
    ["\u5224\u51B3\u6587\u6863\uFF08ops_core_inner_agent\uFF09\u73B0\u8BFB\u4E2D\u2026", "Loading verdict doc (ops_core_inner_agent)\u2026"],
    ["\u5224\u51B3\u6587\u6863\u5C1A\u672A\u751F\u6210\uFF08DocumentDiff/global/", "Verdict doc not generated yet (DocumentDiff/global/"],
    [" \u7F3A\u5E2D\u2014\u2014\u5F15\u64CE\u4EBA\u5BA1\u6302\u8D77\u5373\u81EA\u52A8\u751F\u6210\uFF1B\u7F3A\u5E2D=\u672A\u5230\u751F\u6210\u7A97\u53E3\uFF0C\u4E0D\u7F16\u9020", " absent \u2014 generated automatically on human-approval suspension; absence means the window has not been reached (never fabricated)"],
    ["Diff \u5224\u51B3 \xB7 ops_core_inner_agent", "Diff verdict \xB7 ops_core_inner_agent"],
    ["\u5BF9\u8C61 DocumentDiff/global/", "Object DocumentDiff/global/"],
    /* ── R-UX-10 内容优先批（2026-11-28）── */
    ["\u5185\u5BB9\u4F18\u5148\uFF1A\u6B63\u6587\u4E0E\u76F8\u5173\u4EFB\u52A1\u4F1A\u8BDD\u76F4\u8FBE\uFF1Bscheme \u7ECF\u6309\u94AE\u6309\u9700\u67E5\u770B", "Content first: body & related tasks/sessions up front; scheme on demand via button"],
    ["\u5185\u5BB9\u4F18\u5148\uFF1A\u6B63\u6587/\u76F8\u5173\u4EFB\u52A1\u4E0E\u4F1A\u8BDD/\u6807\u7B7E/\u89C4\u5219/\u64CD\u4F5C\uFF1Bscheme \u5F39\u7A97\u6309\u9700", "Content first: body / related tasks & sessions / tags / rules / ops; scheme on demand in modal"],
    ["\u76F8\u5173\u4EFB\u52A1\u4E0E\u4F1A\u8BDD", "Related tasks & sessions"],
    ["\u65E0\u76F8\u5173\u4EFB\u52A1\u4E0E\u5DE5\u4F5C\u6D41\u8FD0\u884C\uFF08\u5FEB\u7167\u4EFB\u52A1/\u8FD0\u884C\u8BB0\u5F55\u63D0\u53CA\u672C\u5BF9\u8C61\u540D\u96F6\u547D\u4E2D\u2014\u2014\u4E0D\u7F16\u9020\uFF09\u3002", "No related tasks or workflow runs (name not mentioned in task/run snapshots \u2014 nothing fabricated)."],
    ["\u76F8\u5173\u4EFB\u52A1", "Related tasks"],
    ["\u76F8\u5173\u5DE5\u4F5C\u6D41\u8FD0\u884C", "Related workflow runs"],
    ["\u770B\u4F1A\u8BDD", "View session"],
    ["\u4EFB\u52A1\u8BE6\u60C5", "Task detail"],
    ["\u770B\u8FD0\u884C", "View run"],
    ["\u67E5\u770B\u6267\u884C\u4F1A\u8BDD", "View execution session"],
    ["\u5DF2\u6536\u655B", "Converged"],
    ["\u6267\u884C\u4E2D", "Executing"],
    ["\u6302\u8D77", "Suspended"],
    ["scheme \u5B9A\u4E49\uFF08\u5F39\u7A97\u6309\u9700\u67E5\u770B\uFF09", "Scheme definition (view on demand in modal)"],
    ["scheme \xB7 ", "Scheme \xB7 "],
    ["kind \u5B9A\u4E49\uFF08YAML \u7EAF\u6587\u672C\uFF09", "Kind definition (YAML plain text)"],
    ["\u8BE5 kind \u65E0\u5B57\u6BB5\u58F0\u660E\uFF08spec \u7F3A\u5E2D\uFF09\u3002", "This kind declares no spec fields."],
    ["\u65E0\u4ED3\u5E93\u76F4\u8FBE\u58F0\u660E\uFF08connector.download/scope \u8BCD\u5F62\u96F6\u547D\u4E2D\uFF09\u3002", "No repo links declared (no connector.download/scope hits)."],
    ["YAML \u58F0\u660E \xB7 ", "YAML \xB7 "],
    ["\u6267\u884C\u7684\u4EFB\u52A1\uFF08\u672C\u4F1A\u8BDD\u7531\u4EFB\u52A1\u62C9\u8D77\u2014\u2014\u80CC\u666F\u5982\u4E0B\uFF09", "Task behind this session (spawned this session \u2014 background below)"],
    ["\u6765\u6E90\uFF1A", "Source: "],
    ["\u5173\u8054\u4E8B\u4EF6\uFF1A", "Related events: "],
    ["\u5FEB\u7167\u53CD\u67E5\uFF08Task.agentSession \u222A \u4E8B\u4EF6 agentSession\uFF09\u65E0\u5728\u518C\u5173\u8054\u2014\u2014\u4F1A\u8BDD\u578B\u72EC\u7ACB\u4F1A\u8BDD\uFF08\u96F6\u7F16\u9020\uFF09\u3002", "Snapshot reverse lookup (Task.agentSession \u222A events agentSession) found no registered links \u2014 standalone conversation session (nothing fabricated)."],
    /* ── R-UX-10 二次收紧（2026-11-28）：对话修改主通路+详情配置弹窗 ── */
    ["\u5BF9\u8BDD\u4FEE\u6539", "Chat-edit"],
    ["\u5BF9\u8BDD\u4FEE\u6539\uFF08LLM \u6539\u6587\u6863\u2192Apply\u2014\u2014\u4E3B\u901A\u8DEF\uFF09", "Chat-edit (LLM edits the doc \u2192 Apply \u2014 primary path)"],
    ["\u8BF7\u4FEE\u6539\u5DE5\u4F5C\u6D41 ", "Please modify workflow "],
    ["\u8BF7\u4FEE\u6539\u5BF9\u8C61 ", "Please modify object "],
    ["\uFF08YAML \u58F0\u660E\u4E3A\u6743\u5A01\u6E90\uFF09\u3002\u4FEE\u6539\u610F\u56FE\uFF1A", " (YAML declaration is the authoritative source). Intent: "],
    ["\uFF08\u5BF9\u8C61 YAML \u58F0\u660E\u4E3A\u6743\u5A01\u6E90\uFF09\u3002\u4FEE\u6539\u610F\u56FE\uFF1A", " (object YAML declaration is the authoritative source). Intent: "],
    ["\u53D1\u8D77\u5BF9\u8BDD\u4FEE\u6539\uFF08LLM \u6539\u6587\u6863\u2192Apply\u2014\u2014\u4E3B\u901A\u8DEF\uFF09", "Start a chat-edit (LLM edits the doc \u2192 Apply \u2014 primary path)"],
    ["\u914D\u7F6E\u8BE6\u60C5", "Config details"],
    ["\u914D\u7F6E\u8BE6\u60C5\uFF08\u8868\u5355/YAML \u624B\u6539\u2014\u2014\u9ED8\u8BA4\u4E0D\u5C55\u793A\uFF09", "Config details (form/YAML manual edit \u2014 hidden by default)"],
    ["\u914D\u7F6E\u8BE6\u60C5\u7F3A\u5E2D\uFF08\u8BE5 workflow \u65E0\u53EF\u914D\u7F6E studio\uFF09", "Config details unavailable (no configurable studio for this workflow)"],
    ["\u65B0\u5EFA Provider\uFF08\u8868\u5355\u624B\u586B=\u964D\u7EA7\u901A\u8DEF\u2014\u2014\u9ED8\u8BA4\u5BF9\u8BDD\u4FEE\u6539\uFF09", "New Provider (form = fallback path; chat-edit by default)"],
    ["\u6DFB\u52A0 Provider\uFF08OpenAI \u517C\u5BB9\uFF09\u2014\u2014\u624B\u586B\u964D\u7EA7\u901A\u8DEF", "Add Provider (OpenAI-compatible) \u2014 manual fallback path"],
    ["\u9ED8\u8BA4 Provider: ", "Default provider: "],
    ["\u4FEE\u6539\u7ED1\u5B9A", "Edit binding"],
    ["\u4FEE\u6539 Provider \u7ED1\u5B9A \xB7 ", "Edit provider binding \xB7 "],
    /* ── 元数据 header 公用组件（2026-11-28）── */
    ["\u521B\u5EFA ", "Created "],
    ["\u6700\u8FD1\u63D0\u4EA4 ", "Last applied "],
    ["\u91CD\u8981\u5EA6 ", "Importance "],
    ["\u5F52\u5C5E\uFF1A", "Owners: "],
    ["\u5F52\u5C5E\u5BF9\u8C61", "Owning object"],
    ["\u4F1A\u8BDD\u5217\u8868", "Sessions"],
    ["\u4F1A\u8BDD\u6E05\u5355\u5F00\u5408\uFF08\u5177\u4F53\u4F1A\u8BDD\u6DF1\u94FE\u9ED8\u8BA4\u6536\u8D77\uFF09", "Toggle session list (collapsed by default on session deep links)"],
    /* ── 原地对话修改面板 doc-chat（2026-11-28）── */
    ["\u5BF9\u8BDD\u4FEE\u6539 \xB7 ", "Chat-edit \xB7 "],
    ["\u5BF9\u8BDD\u901A\u8DEF\u672A\u6CE8\u5165\uFF08chatfaces.js \u7F3A\u5E2D\uFF09\u2014\u2014\u539F\u5730\u5BF9\u8BDD\u4FEE\u6539\u4E0D\u53EF\u7528\uFF1B\u53EF\u5728\u300C\u539F\u6587\u300D\u89C6\u56FE\u624B\u7F16\uFF0C\u6216\u7ECF\u4F1A\u8BDD\u5C4F\u5BF9\u8BDD\u3002", "Chat bridge not injected (chatfaces.js absent) \u2014 in-place chat-edit unavailable; edit in the raw view or use the chat screen."],
    ["\u5C1A\u65E0\u5BF9\u8BDD\u2014\u2014\u4E0B\u65B9\u8F93\u5165\u4FEE\u6539\u610F\u56FE\u53D1\u9001\u3002", "No messages yet \u2014 type your edit intent below and send."],
    ["\u601D\u8003", "Thinking"],
    ["\u4FEE\u6539\u610F\u56FE\uFF08\u5982\uFF1A\u628A standard \u6A21\u578B\u6362\u6210 reasoning-tier\uFF09\u2014\u2014\u53D1\u9001\u540E\u52A9\u624B\u56DE\u4FEE\u6539\u540E YAML\uFF0C\u5BF9\u6BD4\u786E\u8BA4\u5373 Apply", "Edit intent (e.g. switch standard model to reasoning-tier) \u2014 the assistant replies with modified YAML; compare then Apply"],
    ["\u5BF9\u6BD4\u5E76 Apply", "Compare & Apply"],
    ["\u53D1\u9001", "Send"],
    ["\u505C\u6B62", "Stop"],
    ["\u8BF7\u5148\u5199\u4FEE\u6539\u610F\u56FE", "Write your edit intent first"],
    ["\u4E0A\u4E00\u8F6E\u4ECD\u5728\u8FD0\u884C\u2014\u2014\u53EF\u70B9\u505C\u6B62", "Previous turn still running \u2014 you can stop it"],
    ["\u53D1\u9001\u5931\u8D25\uFF1A", "Send failed: "],
    ["\u65B0\u4F1A\u8BDD\u5931\u8D25\uFF08", "New session failed ("],
    ["\u6700\u8FD1\u56DE\u590D\u672A\u542B ```yaml \u4EE3\u7801\u5757\u2014\u2014\u8BF7\u8BA9\u52A9\u624B\u7ED9\u51FA\u4FEE\u6539\u540E\u7684\u5B8C\u6574 YAML \u58F0\u660E\u3002", "The latest reply has no ```yaml block \u2014 ask the assistant for the full modified YAML declaration."],
    ["\u4FEE\u6539\u9884\u89C8\uFF08", "Change preview ("],
    [" \u5904\u53D8\u66F4\uFF09", " changes)"],
    ["\u6279\u51C6\uFF08Apply\uFF09", "Approve (Apply)"],
    ["\u653E\u5F03", "Discard"],
    ["\u6B63\u5728 Apply\u2026", "Applying\u2026"],
    ["\u5DF2 Apply\uFF08rv=", "Applied (rv="],
    ["\u5BF9\u8BDD\u4FEE\u6539\u5DF2 Apply", "Chat edit applied"],
    ["Apply \u5931\u8D25\uFF1A", "Apply failed: "],
    ["\uFF08\u7248\u672C\u5DF2\u53D8\u2014\u2014\u91CD\u8BD5\u5C06\u53D6\u65B0\u7248\u672C\uFF09", " (version changed \u2014 retry fetches the new one)"],
    ["\u8BF7\u4FEE\u6539\u5BF9\u8C61 ", "Please modify object "],
    ["\uFF08\u5F53\u524D YAML \u58F0\u660E\u4E3A\u6743\u5A01\u6E90\uFF0C\u5168\u6587\u5982\u4E0B\uFF09\u3002\u4FEE\u6539\u610F\u56FE\uFF1A", " (current YAML declaration below is authoritative). Edit intent: "],
    ["\u8F93\u51FA\u8981\u6C42\uFF1A\u7ED9\u51FA\u4FEE\u6539\u540E\u7684\u5B8C\u6574 YAML \u58F0\u660E\uFF0C\u653E\u5728\u4E00\u4E2A ```yaml \u4EE3\u7801\u5757\u91CC\uFF1B\u4E0D\u6539\u65E0\u5173\u5B57\u6BB5\uFF1B\u89E3\u91CA\u653E\u4EE3\u7801\u5757\u5916\u3002", "Output: the complete modified YAML declaration in one ```yaml block; do not touch unrelated fields; put explanations outside the block."]
  ];

  // src/core/i18n.js
  var KEY = "cc.locale";
  var locale = "zh";
  try {
    locale = typeof localStorage !== "undefined" && localStorage.getItem(KEY) || "zh";
  } catch {
  }
  var en = new Map(enEntries);
  var DICTS = { en };
  var subs = /* @__PURE__ */ new Set();
  function t(s) {
    if (locale === "zh") return String(s);
    const d = DICTS[locale];
    return d && d.get(String(s)) || String(s);
  }
  function getLocale() {
    return locale;
  }
  function setLocale(l) {
    if (l !== "zh" && !DICTS[l]) return;
    locale = l;
    try {
      if (typeof localStorage !== "undefined") localStorage.setItem(KEY, l);
    } catch {
    }
    subs.forEach((f) => {
      try {
        f(l);
      } catch {
      }
    });
  }
  function onLocaleChange(f) {
    subs.add(f);
    return () => subs.delete(f);
  }

  // src/core/format.js
  function esc(s) {
    return String(s == null ? "" : s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#39;");
  }
  function safeHref(u) {
    const s = String(u || "").trim();
    if (!s) return "";
    if (/^https?:\/\//i.test(s)) return s;
    if (/^(#|\/|\.\.?\/)/.test(s)) return s;
    return "";
  }
  function pad(n) {
    return n < 10 ? "0" + n : "" + n;
  }
  function toDate(v) {
    if (v instanceof Date) return v;
    if (typeof v === "number") return new Date(v);
    return new Date(String(v == null ? "" : v).replace(" ", "T"));
  }
  function fmtTime(v) {
    const d = toDate(v);
    return isNaN(d) ? "\u2014" : pad(d.getHours()) + ":" + pad(d.getMinutes());
  }
  function fmtTimeS(v) {
    const d = toDate(v);
    return isNaN(d) ? "\u2014" : fmtTime(d) + ":" + pad(d.getSeconds());
  }
  function fmtDate(v) {
    const d = toDate(v);
    return isNaN(d) ? "\u2014" : d.getFullYear() + "-" + pad(d.getMonth() + 1) + "-" + pad(d.getDate());
  }
  function fmtAbs(v) {
    const d = toDate(v);
    if (isNaN(d)) return "\u65F6\u95F4\u672A\u77E5";
    let zone = "";
    try {
      zone = Intl.DateTimeFormat().resolvedOptions().timeZone || "";
    } catch {
      zone = "";
    }
    return fmtDate(d) + " " + fmtTimeS(d) + (zone ? " " + zone : "");
  }
  function relAge(ts, nowMs) {
    const t2 = toDate(ts).getTime();
    if (isNaN(t2)) return "";
    const d = Math.max(0, Math.round(((nowMs || Date.now()) - t2) / 1e3));
    const en2 = getLocale() === "en";
    const unit = (n, zh, e) => en2 ? n + e : n + zh;
    if (d < 60) return unit(d, "s \u524D", "s ago");
    if (d < 3600) return unit(Math.round(d / 60), "m \u524D", "m ago");
    if (d < 86400) return unit(Math.round(d / 3600), "h \u524D", "h ago");
    return unit(Math.round(d / 86400), "d \u524D", "d ago");
  }
  function timeHtml(value) {
    const d = toDate(value);
    const iso = isNaN(d) ? "" : d.toISOString();
    return '<time class="mono" datetime="' + esc(iso) + '" title="' + esc(fmtAbs(d)) + '">' + esc(relAge(value) || "\u65F6\u95F4\u672A\u77E5") + "</time>";
  }
  function fmtNum(n) {
    if (n == null || isNaN(Number(n))) return "\u2014";
    return Number(n).toLocaleString("zh-CN", { maximumFractionDigits: 1 });
  }
  function trunc(s, n) {
    const t2 = String(s == null ? "" : s);
    return t2.length > n ? t2.slice(0, n) + "\u2026" : t2;
  }

  // ../../../web/node_modules/fast-json-patch/module/core.mjs
  var core_exports = {};
  __export(core_exports, {
    JsonPatchError: () => JsonPatchError,
    _areEquals: () => _areEquals,
    applyOperation: () => applyOperation,
    applyPatch: () => applyPatch,
    applyReducer: () => applyReducer,
    deepClone: () => deepClone,
    getValueByPointer: () => getValueByPointer,
    validate: () => validate,
    validator: () => validator
  });

  // ../../../web/node_modules/fast-json-patch/module/helpers.mjs
  var __extends = /* @__PURE__ */ (function() {
    var extendStatics = function(d, b) {
      extendStatics = Object.setPrototypeOf || { __proto__: [] } instanceof Array && function(d2, b2) {
        d2.__proto__ = b2;
      } || function(d2, b2) {
        for (var p in b2) if (b2.hasOwnProperty(p)) d2[p] = b2[p];
      };
      return extendStatics(d, b);
    };
    return function(d, b) {
      extendStatics(d, b);
      function __() {
        this.constructor = d;
      }
      d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
  })();
  var _hasOwnProperty = Object.prototype.hasOwnProperty;
  function hasOwnProperty(obj, key) {
    return _hasOwnProperty.call(obj, key);
  }
  function _objectKeys(obj) {
    if (Array.isArray(obj)) {
      var keys_1 = new Array(obj.length);
      for (var k = 0; k < keys_1.length; k++) {
        keys_1[k] = "" + k;
      }
      return keys_1;
    }
    if (Object.keys) {
      return Object.keys(obj);
    }
    var keys = [];
    for (var i in obj) {
      if (hasOwnProperty(obj, i)) {
        keys.push(i);
      }
    }
    return keys;
  }
  function _deepClone(obj) {
    switch (typeof obj) {
      case "object":
        return JSON.parse(JSON.stringify(obj));
      //Faster than ES5 clone - http://jsperf.com/deep-cloning-of-objects/5
      case "undefined":
        return null;
      //this is how JSON.stringify behaves for array items
      default:
        return obj;
    }
  }
  function isInteger(str2) {
    var i = 0;
    var len = str2.length;
    var charCode;
    while (i < len) {
      charCode = str2.charCodeAt(i);
      if (charCode >= 48 && charCode <= 57) {
        i++;
        continue;
      }
      return false;
    }
    return true;
  }
  function escapePathComponent(path) {
    if (path.indexOf("/") === -1 && path.indexOf("~") === -1)
      return path;
    return path.replace(/~/g, "~0").replace(/\//g, "~1");
  }
  function unescapePathComponent(path) {
    return path.replace(/~1/g, "/").replace(/~0/g, "~");
  }
  function hasUndefined(obj) {
    if (obj === void 0) {
      return true;
    }
    if (obj) {
      if (Array.isArray(obj)) {
        for (var i_1 = 0, len = obj.length; i_1 < len; i_1++) {
          if (hasUndefined(obj[i_1])) {
            return true;
          }
        }
      } else if (typeof obj === "object") {
        var objKeys = _objectKeys(obj);
        var objKeysLength = objKeys.length;
        for (var i = 0; i < objKeysLength; i++) {
          if (hasUndefined(obj[objKeys[i]])) {
            return true;
          }
        }
      }
    }
    return false;
  }
  function patchErrorMessageFormatter(message, args) {
    var messageParts = [message];
    for (var key in args) {
      var value = typeof args[key] === "object" ? JSON.stringify(args[key], null, 2) : args[key];
      if (typeof value !== "undefined") {
        messageParts.push(key + ": " + value);
      }
    }
    return messageParts.join("\n");
  }
  var PatchError = (
    /** @class */
    (function(_super) {
      __extends(PatchError2, _super);
      function PatchError2(message, name, index, operation, tree) {
        var _newTarget = this.constructor;
        var _this = _super.call(this, patchErrorMessageFormatter(message, { name, index, operation, tree })) || this;
        _this.name = name;
        _this.index = index;
        _this.operation = operation;
        _this.tree = tree;
        Object.setPrototypeOf(_this, _newTarget.prototype);
        _this.message = patchErrorMessageFormatter(message, { name, index, operation, tree });
        return _this;
      }
      return PatchError2;
    })(Error)
  );

  // ../../../web/node_modules/fast-json-patch/module/core.mjs
  var JsonPatchError = PatchError;
  var deepClone = _deepClone;
  var objOps = {
    add: function(obj, key, document2) {
      obj[key] = this.value;
      return { newDocument: document2 };
    },
    remove: function(obj, key, document2) {
      var removed = obj[key];
      delete obj[key];
      return { newDocument: document2, removed };
    },
    replace: function(obj, key, document2) {
      var removed = obj[key];
      obj[key] = this.value;
      return { newDocument: document2, removed };
    },
    move: function(obj, key, document2) {
      var removed = getValueByPointer(document2, this.path);
      if (removed) {
        removed = _deepClone(removed);
      }
      var originalValue = applyOperation(document2, { op: "remove", path: this.from }).removed;
      applyOperation(document2, { op: "add", path: this.path, value: originalValue });
      return { newDocument: document2, removed };
    },
    copy: function(obj, key, document2) {
      var valueToCopy = getValueByPointer(document2, this.from);
      applyOperation(document2, { op: "add", path: this.path, value: _deepClone(valueToCopy) });
      return { newDocument: document2 };
    },
    test: function(obj, key, document2) {
      return { newDocument: document2, test: _areEquals(obj[key], this.value) };
    },
    _get: function(obj, key, document2) {
      this.value = obj[key];
      return { newDocument: document2 };
    }
  };
  var arrOps = {
    add: function(arr, i, document2) {
      if (isInteger(i)) {
        arr.splice(i, 0, this.value);
      } else {
        arr[i] = this.value;
      }
      return { newDocument: document2, index: i };
    },
    remove: function(arr, i, document2) {
      var removedList = arr.splice(i, 1);
      return { newDocument: document2, removed: removedList[0] };
    },
    replace: function(arr, i, document2) {
      var removed = arr[i];
      arr[i] = this.value;
      return { newDocument: document2, removed };
    },
    move: objOps.move,
    copy: objOps.copy,
    test: objOps.test,
    _get: objOps._get
  };
  function getValueByPointer(document2, pointer) {
    if (pointer == "") {
      return document2;
    }
    var getOriginalDestination = { op: "_get", path: pointer };
    applyOperation(document2, getOriginalDestination);
    return getOriginalDestination.value;
  }
  function applyOperation(document2, operation, validateOperation, mutateDocument, banPrototypeModifications, index) {
    if (validateOperation === void 0) {
      validateOperation = false;
    }
    if (mutateDocument === void 0) {
      mutateDocument = true;
    }
    if (banPrototypeModifications === void 0) {
      banPrototypeModifications = true;
    }
    if (index === void 0) {
      index = 0;
    }
    if (validateOperation) {
      if (typeof validateOperation == "function") {
        validateOperation(operation, 0, document2, operation.path);
      } else {
        validator(operation, 0);
      }
    }
    if (operation.path === "") {
      var returnValue = { newDocument: document2 };
      if (operation.op === "add") {
        returnValue.newDocument = operation.value;
        return returnValue;
      } else if (operation.op === "replace") {
        returnValue.newDocument = operation.value;
        returnValue.removed = document2;
        return returnValue;
      } else if (operation.op === "move" || operation.op === "copy") {
        returnValue.newDocument = getValueByPointer(document2, operation.from);
        if (operation.op === "move") {
          returnValue.removed = document2;
        }
        return returnValue;
      } else if (operation.op === "test") {
        returnValue.test = _areEquals(document2, operation.value);
        if (returnValue.test === false) {
          throw new JsonPatchError("Test operation failed", "TEST_OPERATION_FAILED", index, operation, document2);
        }
        returnValue.newDocument = document2;
        return returnValue;
      } else if (operation.op === "remove") {
        returnValue.removed = document2;
        returnValue.newDocument = null;
        return returnValue;
      } else if (operation.op === "_get") {
        operation.value = document2;
        return returnValue;
      } else {
        if (validateOperation) {
          throw new JsonPatchError("Operation `op` property is not one of operations defined in RFC-6902", "OPERATION_OP_INVALID", index, operation, document2);
        } else {
          return returnValue;
        }
      }
    } else {
      if (!mutateDocument) {
        document2 = _deepClone(document2);
      }
      var path = operation.path || "";
      var keys = path.split("/");
      var obj = document2;
      var t2 = 1;
      var len = keys.length;
      var existingPathFragment = void 0;
      var key = void 0;
      var validateFunction = void 0;
      if (typeof validateOperation == "function") {
        validateFunction = validateOperation;
      } else {
        validateFunction = validator;
      }
      while (true) {
        key = keys[t2];
        if (key && key.indexOf("~") != -1) {
          key = unescapePathComponent(key);
        }
        if (banPrototypeModifications && (key == "__proto__" || key == "prototype" && t2 > 0 && keys[t2 - 1] == "constructor")) {
          throw new TypeError("JSON-Patch: modifying `__proto__` or `constructor/prototype` prop is banned for security reasons, if this was on purpose, please set `banPrototypeModifications` flag false and pass it to this function. More info in fast-json-patch README");
        }
        if (validateOperation) {
          if (existingPathFragment === void 0) {
            if (obj[key] === void 0) {
              existingPathFragment = keys.slice(0, t2).join("/");
            } else if (t2 == len - 1) {
              existingPathFragment = operation.path;
            }
            if (existingPathFragment !== void 0) {
              validateFunction(operation, 0, document2, existingPathFragment);
            }
          }
        }
        t2++;
        if (Array.isArray(obj)) {
          if (key === "-") {
            key = obj.length;
          } else {
            if (validateOperation && !isInteger(key)) {
              throw new JsonPatchError("Expected an unsigned base-10 integer value, making the new referenced value the array element with the zero-based index", "OPERATION_PATH_ILLEGAL_ARRAY_INDEX", index, operation, document2);
            } else if (isInteger(key)) {
              key = ~~key;
            }
          }
          if (t2 >= len) {
            if (validateOperation && operation.op === "add" && key > obj.length) {
              throw new JsonPatchError("The specified index MUST NOT be greater than the number of elements in the array", "OPERATION_VALUE_OUT_OF_BOUNDS", index, operation, document2);
            }
            var returnValue = arrOps[operation.op].call(operation, obj, key, document2);
            if (returnValue.test === false) {
              throw new JsonPatchError("Test operation failed", "TEST_OPERATION_FAILED", index, operation, document2);
            }
            return returnValue;
          }
        } else {
          if (t2 >= len) {
            var returnValue = objOps[operation.op].call(operation, obj, key, document2);
            if (returnValue.test === false) {
              throw new JsonPatchError("Test operation failed", "TEST_OPERATION_FAILED", index, operation, document2);
            }
            return returnValue;
          }
        }
        obj = obj[key];
        if (validateOperation && t2 < len && (!obj || typeof obj !== "object")) {
          throw new JsonPatchError("Cannot perform operation at the desired path", "OPERATION_PATH_UNRESOLVABLE", index, operation, document2);
        }
      }
    }
  }
  function applyPatch(document2, patch, validateOperation, mutateDocument, banPrototypeModifications) {
    if (mutateDocument === void 0) {
      mutateDocument = true;
    }
    if (banPrototypeModifications === void 0) {
      banPrototypeModifications = true;
    }
    if (validateOperation) {
      if (!Array.isArray(patch)) {
        throw new JsonPatchError("Patch sequence must be an array", "SEQUENCE_NOT_AN_ARRAY");
      }
    }
    if (!mutateDocument) {
      document2 = _deepClone(document2);
    }
    var results = new Array(patch.length);
    for (var i = 0, length_1 = patch.length; i < length_1; i++) {
      results[i] = applyOperation(document2, patch[i], validateOperation, true, banPrototypeModifications, i);
      document2 = results[i].newDocument;
    }
    results.newDocument = document2;
    return results;
  }
  function applyReducer(document2, operation, index) {
    var operationResult = applyOperation(document2, operation);
    if (operationResult.test === false) {
      throw new JsonPatchError("Test operation failed", "TEST_OPERATION_FAILED", index, operation, document2);
    }
    return operationResult.newDocument;
  }
  function validator(operation, index, document2, existingPathFragment) {
    if (typeof operation !== "object" || operation === null || Array.isArray(operation)) {
      throw new JsonPatchError("Operation is not an object", "OPERATION_NOT_AN_OBJECT", index, operation, document2);
    } else if (!objOps[operation.op]) {
      throw new JsonPatchError("Operation `op` property is not one of operations defined in RFC-6902", "OPERATION_OP_INVALID", index, operation, document2);
    } else if (typeof operation.path !== "string") {
      throw new JsonPatchError("Operation `path` property is not a string", "OPERATION_PATH_INVALID", index, operation, document2);
    } else if (operation.path.indexOf("/") !== 0 && operation.path.length > 0) {
      throw new JsonPatchError('Operation `path` property must start with "/"', "OPERATION_PATH_INVALID", index, operation, document2);
    } else if ((operation.op === "move" || operation.op === "copy") && typeof operation.from !== "string") {
      throw new JsonPatchError("Operation `from` property is not present (applicable in `move` and `copy` operations)", "OPERATION_FROM_REQUIRED", index, operation, document2);
    } else if ((operation.op === "add" || operation.op === "replace" || operation.op === "test") && operation.value === void 0) {
      throw new JsonPatchError("Operation `value` property is not present (applicable in `add`, `replace` and `test` operations)", "OPERATION_VALUE_REQUIRED", index, operation, document2);
    } else if ((operation.op === "add" || operation.op === "replace" || operation.op === "test") && hasUndefined(operation.value)) {
      throw new JsonPatchError("Operation `value` property is not present (applicable in `add`, `replace` and `test` operations)", "OPERATION_VALUE_CANNOT_CONTAIN_UNDEFINED", index, operation, document2);
    } else if (document2) {
      if (operation.op == "add") {
        var pathLen = operation.path.split("/").length;
        var existingPathLen = existingPathFragment.split("/").length;
        if (pathLen !== existingPathLen + 1 && pathLen !== existingPathLen) {
          throw new JsonPatchError("Cannot perform an `add` operation at the desired path", "OPERATION_PATH_CANNOT_ADD", index, operation, document2);
        }
      } else if (operation.op === "replace" || operation.op === "remove" || operation.op === "_get") {
        if (operation.path !== existingPathFragment) {
          throw new JsonPatchError("Cannot perform the operation at a path that does not exist", "OPERATION_PATH_UNRESOLVABLE", index, operation, document2);
        }
      } else if (operation.op === "move" || operation.op === "copy") {
        var existingValue = { op: "_get", path: operation.from, value: void 0 };
        var error = validate([existingValue], document2);
        if (error && error.name === "OPERATION_PATH_UNRESOLVABLE") {
          throw new JsonPatchError("Cannot perform the operation from a path that does not exist", "OPERATION_FROM_UNRESOLVABLE", index, operation, document2);
        }
      }
    }
  }
  function validate(sequence, document2, externalValidator) {
    try {
      if (!Array.isArray(sequence)) {
        throw new JsonPatchError("Patch sequence must be an array", "SEQUENCE_NOT_AN_ARRAY");
      }
      if (document2) {
        applyPatch(_deepClone(document2), _deepClone(sequence), externalValidator || true);
      } else {
        externalValidator = externalValidator || validator;
        for (var i = 0; i < sequence.length; i++) {
          externalValidator(sequence[i], i, document2, void 0);
        }
      }
    } catch (e) {
      if (e instanceof JsonPatchError) {
        return e;
      } else {
        throw e;
      }
    }
  }
  function _areEquals(a, b) {
    if (a === b)
      return true;
    if (a && b && typeof a == "object" && typeof b == "object") {
      var arrA = Array.isArray(a), arrB = Array.isArray(b), i, length, key;
      if (arrA && arrB) {
        length = a.length;
        if (length != b.length)
          return false;
        for (i = length; i-- !== 0; )
          if (!_areEquals(a[i], b[i]))
            return false;
        return true;
      }
      if (arrA != arrB)
        return false;
      var keys = Object.keys(a);
      length = keys.length;
      if (length !== Object.keys(b).length)
        return false;
      for (i = length; i-- !== 0; )
        if (!b.hasOwnProperty(keys[i]))
          return false;
      for (i = length; i-- !== 0; ) {
        key = keys[i];
        if (!_areEquals(a[key], b[key]))
          return false;
      }
      return true;
    }
    return a !== a && b !== b;
  }

  // ../../../web/node_modules/fast-json-patch/module/duplex.mjs
  var duplex_exports = {};
  __export(duplex_exports, {
    compare: () => compare,
    generate: () => generate,
    observe: () => observe,
    unobserve: () => unobserve
  });
  var beforeDict = /* @__PURE__ */ new WeakMap();
  var Mirror = (
    /** @class */
    /* @__PURE__ */ (function() {
      function Mirror2(obj) {
        this.observers = /* @__PURE__ */ new Map();
        this.obj = obj;
      }
      return Mirror2;
    })()
  );
  var ObserverInfo = (
    /** @class */
    /* @__PURE__ */ (function() {
      function ObserverInfo2(callback, observer) {
        this.callback = callback;
        this.observer = observer;
      }
      return ObserverInfo2;
    })()
  );
  function getMirror(obj) {
    return beforeDict.get(obj);
  }
  function getObserverFromMirror(mirror, callback) {
    return mirror.observers.get(callback);
  }
  function removeObserverFromMirror(mirror, observer) {
    mirror.observers.delete(observer.callback);
  }
  function unobserve(root, observer) {
    observer.unobserve();
  }
  function observe(obj, callback) {
    var patches = [];
    var observer;
    var mirror = getMirror(obj);
    if (!mirror) {
      mirror = new Mirror(obj);
      beforeDict.set(obj, mirror);
    } else {
      var observerInfo = getObserverFromMirror(mirror, callback);
      observer = observerInfo && observerInfo.observer;
    }
    if (observer) {
      return observer;
    }
    observer = {};
    mirror.value = _deepClone(obj);
    if (callback) {
      observer.callback = callback;
      observer.next = null;
      var dirtyCheck = function() {
        generate(observer);
      };
      var fastCheck = function() {
        clearTimeout(observer.next);
        observer.next = setTimeout(dirtyCheck);
      };
      if (typeof window !== "undefined") {
        window.addEventListener("mouseup", fastCheck);
        window.addEventListener("keyup", fastCheck);
        window.addEventListener("mousedown", fastCheck);
        window.addEventListener("keydown", fastCheck);
        window.addEventListener("change", fastCheck);
      }
    }
    observer.patches = patches;
    observer.object = obj;
    observer.unobserve = function() {
      generate(observer);
      clearTimeout(observer.next);
      removeObserverFromMirror(mirror, observer);
      if (typeof window !== "undefined") {
        window.removeEventListener("mouseup", fastCheck);
        window.removeEventListener("keyup", fastCheck);
        window.removeEventListener("mousedown", fastCheck);
        window.removeEventListener("keydown", fastCheck);
        window.removeEventListener("change", fastCheck);
      }
    };
    mirror.observers.set(callback, new ObserverInfo(callback, observer));
    return observer;
  }
  function generate(observer, invertible) {
    if (invertible === void 0) {
      invertible = false;
    }
    var mirror = beforeDict.get(observer.object);
    _generate(mirror.value, observer.object, observer.patches, "", invertible);
    if (observer.patches.length) {
      applyPatch(mirror.value, observer.patches);
    }
    var temp = observer.patches;
    if (temp.length > 0) {
      observer.patches = [];
      if (observer.callback) {
        observer.callback(temp);
      }
    }
    return temp;
  }
  function _generate(mirror, obj, patches, path, invertible) {
    if (obj === mirror) {
      return;
    }
    if (typeof obj.toJSON === "function") {
      obj = obj.toJSON();
    }
    var newKeys = _objectKeys(obj);
    var oldKeys = _objectKeys(mirror);
    var changed = false;
    var deleted = false;
    for (var t2 = oldKeys.length - 1; t2 >= 0; t2--) {
      var key = oldKeys[t2];
      var oldVal = mirror[key];
      if (hasOwnProperty(obj, key) && !(obj[key] === void 0 && oldVal !== void 0 && Array.isArray(obj) === false)) {
        var newVal = obj[key];
        if (typeof oldVal == "object" && oldVal != null && typeof newVal == "object" && newVal != null && Array.isArray(oldVal) === Array.isArray(newVal)) {
          _generate(oldVal, newVal, patches, path + "/" + escapePathComponent(key), invertible);
        } else {
          if (oldVal !== newVal) {
            changed = true;
            if (invertible) {
              patches.push({ op: "test", path: path + "/" + escapePathComponent(key), value: _deepClone(oldVal) });
            }
            patches.push({ op: "replace", path: path + "/" + escapePathComponent(key), value: _deepClone(newVal) });
          }
        }
      } else if (Array.isArray(mirror) === Array.isArray(obj)) {
        if (invertible) {
          patches.push({ op: "test", path: path + "/" + escapePathComponent(key), value: _deepClone(oldVal) });
        }
        patches.push({ op: "remove", path: path + "/" + escapePathComponent(key) });
        deleted = true;
      } else {
        if (invertible) {
          patches.push({ op: "test", path, value: mirror });
        }
        patches.push({ op: "replace", path, value: obj });
        changed = true;
      }
    }
    if (!deleted && newKeys.length == oldKeys.length) {
      return;
    }
    for (var t2 = 0; t2 < newKeys.length; t2++) {
      var key = newKeys[t2];
      if (!hasOwnProperty(mirror, key) && obj[key] !== void 0) {
        patches.push({ op: "add", path: path + "/" + escapePathComponent(key), value: _deepClone(obj[key]) });
      }
    }
  }
  function compare(tree1, tree2, invertible) {
    if (invertible === void 0) {
      invertible = false;
    }
    var patches = [];
    _generate(tree1, tree2, patches, "", invertible);
    return patches;
  }

  // ../../../web/node_modules/fast-json-patch/index.mjs
  var fast_json_patch_default = Object.assign({}, core_exports, duplex_exports, {
    JsonPatchError: PatchError,
    deepClone: _deepClone,
    escapePathComponent,
    unescapePathComponent
  });

  // src/core/store.js
  var { applyPatch: applyPatch2 } = fast_json_patch_default;
  function objKey(kind, scope, name) {
    return kind + "|" + (scope || "global") + "|" + (name || "");
  }
  function normalizeRef(x, kindHint) {
    if (x == null) return null;
    if (typeof x === "string") {
      return kindHint ? { kind: String(kindHint), scope: "global", name: x, rv: 0, raw: x } : null;
    }
    if (typeof x !== "object") return null;
    const kind = x.kind || kindHint;
    const name = x.name;
    if (!kind || !name) return null;
    const rvRaw = x.resourceVersion != null ? x.resourceVersion : x.rv;
    return { kind: String(kind), scope: String(x.scope || "global"), name: String(name), rv: Number(rvRaw) || 0, raw: x };
  }
  function arrOf(v) {
    return Array.isArray(v) ? v : [];
  }
  function cloneSnapshot(v) {
    try {
      return typeof structuredClone === "function" ? structuredClone(v) : JSON.parse(JSON.stringify(v));
    } catch {
      return v;
    }
  }
  function createStore() {
    let state = null;
    let seq2 = 0;
    const subs2 = /* @__PURE__ */ new Set();
    let scheduled = false;
    const docCache = /* @__PURE__ */ new Map();
    function notify() {
      if (scheduled) return;
      scheduled = true;
      queueMicrotask(() => {
        scheduled = false;
        subs2.forEach((fn) => {
          try {
            fn(state);
          } catch (e) {
            console.error("[store] \u8BA2\u9605\u8005\u5931\u8D25", e);
          }
        });
      });
    }
    const api2 = {
      /** 快照是否就绪（未 init 前=骨架词形）。 */
      get isReady() {
        return !!state;
      },
      /** 建基线（整段快照——首连/patch 失败回退/降级轮询共用同一入口）。 */
      init(snapshot) {
        state = snapshot && typeof snapshot === "object" ? cloneSnapshot(snapshot) : {};
        if (Number.isFinite(Number(state.seq))) seq2 = Number(state.seq);
        notify();
        return api2;
      },
      /**
       * 增量应用 RFC6902 patch（validate:false 保守应用）。
       * 返回 true=已应用并通知；false=应用失败（状态保持原样——调用方整段 re-fetch）。
       */
      applyPatch(patch, meta) {
        if (!state) return false;
        if (!Array.isArray(patch) || !patch.length) return true;
        let ok = true;
        try {
          const res = applyPatch2(state, patch, false, true);
          if (res && res.newDocument) state = res.newDocument;
        } catch (e) {
          ok = false;
          console.warn("[store] patch \u5E94\u7528\u5931\u8D25\u2014\u2014\u56DE\u9000\u6574\u6BB5\u5FEB\u7167\uFF1A", String(e && e.message || e));
        }
        if (!ok) return false;
        if (meta && Number.isFinite(Number(meta.seq))) seq2 = Number(meta.seq);
        notify();
        return true;
      },
      /** 最近应用的 patch 序号。 */
      get seq() {
        return seq2;
      },
      /** 连接态（connecting/connected/reconnecting/degraded——ws.js 广播，随状态帧通知）。 */
      setConn(s) {
        if (conn !== s) {
          conn = s;
          notify();
        }
      },
      get conn() {
        return conn;
      },
      /** 点路径读（'a.b.0.c'——缺路径=undefined 不抛）。 */
      get(path) {
        if (!state) return void 0;
        return String(path || "").split(".").reduce((o, k) => o == null ? void 0 : o[k], state);
      },
      /** 订阅状态帧（返回退订函数）。 */
      subscribe(fn) {
        subs2.add(fn);
        return () => subs2.delete(fn);
      },
      /* ---------- 容器化访问器（快照字段缺席=空数组/空值，不抛不猜） ---------- */
      objects() {
        const o = state && state.objects;
        if (!o) return [];
        if (Array.isArray(o)) return o.map((x) => normalizeRef(x)).filter(Boolean);
        return Object.keys(o).flatMap((k) => arrOf(o[k]).map((x) => normalizeRef(x, k))).filter(Boolean);
      },
      refs(kind) {
        return api2.objects().filter((x) => x.kind === kind);
      },
      kinds() {
        return arrOf(state && state.kinds);
      },
      events() {
        return arrOf(state && state.events);
      },
      drafts() {
        return arrOf(state && state.drafts);
      },
      tasks() {
        const t2 = state && state.tasks;
        if (Array.isArray(t2)) return t2;
        return arrOf(t2 && t2.tasks);
      },
      runs() {
        const r = state && (state.runs !== void 0 ? state.runs : state.workflowRuns);
        if (Array.isArray(r)) return r;
        if (r && Array.isArray(r.recent)) return r.recent;
        return arrOf(r && r.runs);
      },
      workflows() {
        const w = state && state.workflows;
        return Array.isArray(w) ? w : arrOf(w && w.workflows);
      },
      proposals() {
        const p = state && state.proposals;
        return Array.isArray(p) ? p : arrOf(p && p.proposals);
      },
      sessions() {
        const s = state && state.sessions;
        return Array.isArray(s) ? s : arrOf(s && s.sessions);
      },
      attention() {
        const a = state && state.attention;
        if (!a) return [];
        if (Array.isArray(a)) return a;
        return arrOf(a.items || a.cards);
      },
      inbox() {
        return state && state.inbox || null;
      },
      deliveries() {
        return state && state.deliveries || null;
      },
      codeAudit() {
        return state && state.codeAudit || null;
      },
      /* ---------- 对象级 rv 缓存（正文懒取的家） ---------- */
      cacheDoc(kind, scope, name, rv, content) {
        docCache.set(objKey(kind, scope, name), {
          rv: Number(rv) || 0,
          content: content == null ? "" : String(content),
          at: Date.now()
        });
      },
      /** 命中判据：rv 未变才返回缓存（rv 变了=过期不返回——调用方现读）。 */
      cachedDoc(kind, scope, name, rv) {
        const c = docCache.get(objKey(kind, scope, name));
        if (!c) return null;
        const want = Number(rv) || 0;
        if (want > 0 && c.rv !== want) return null;
        return c;
      },
      /** 快照中 rv 与缓存不符（或未缓存）的对象 refs——首连/降级轮询后正文预取清单。 */
      staleDocs(limit) {
        const cap = Number(limit) || 24;
        const out = [];
        for (const r of api2.objects()) {
          if (!r.kind || !r.name) continue;
          const c = docCache.get(objKey(r.kind, r.scope, r.name));
          if (!c || c.rv !== r.rv) {
            out.push(r);
            if (out.length >= cap) break;
          }
        }
        return out;
      },
      clearDocCache() {
        docCache.clear();
      }
    };
    let conn = "connecting";
    return api2;
  }

  // src/core/api.js
  function createApi(base = "", paths = {}) {
    const B = String(base || "").replace(/\/+$/, "");
    const enc = encodeURIComponent;
    const P = {
      apply: paths.apply || "v1/apply",
      drafts: paths.drafts || "v1/drafts"
    };
    async function request(path, opts = {}) {
      const p = String(path);
      const url = p.startsWith("/") ? p : B + "/" + p.replace(/^\/+/, "");
      const timeoutMs = opts.timeout != null ? Number(opts.timeout) : opts.method && opts.method !== "GET" ? 3e4 : 15e3;
      const ctl = typeof AbortController !== "undefined" ? new AbortController() : null;
      const timer = ctl && timeoutMs > 0 ? setTimeout(() => ctl.abort(), timeoutMs) : null;
      try {
        const res = await fetch(url, {
          method: opts.method || "GET",
          headers: opts.body !== void 0 ? { "Content-Type": "application/json" } : void 0,
          body: opts.body !== void 0 ? JSON.stringify(opts.body) : void 0,
          credentials: "same-origin",
          signal: ctl ? ctl.signal : void 0
        });
        const text2 = await res.text();
        let data = null;
        try {
          data = text2 ? JSON.parse(text2) : null;
        } catch {
          data = { raw: text2 };
        }
        if (!res.ok) {
          const err = data && (data.error || data.message);
          return { ok: false, status: res.status, error: err ? String(err) : "HTTP " + res.status };
        }
        return { ok: true, data };
      } catch (e) {
        const aborted = ctl && ctl.signal.aborted;
        return {
          ok: false,
          status: 0,
          error: aborted ? "\u8BF7\u6C42\u8D85\u65F6\uFF08\u540E\u7AEF\u672A\u5728 " + timeoutMs + "ms \u5185\u54CD\u5E94\uFF0C\u6062\u590D\u540E\u91CD\u8FDB\u5373\u73B0\u8BFB\uFF09" : String(e && e.message || e)
        };
      } finally {
        if (timer) clearTimeout(timer);
      }
    }
    return {
      request,
      /* ---------- 新契约端点 ---------- */
      snapshot: () => request("v1/snapshot"),
      event: (id) => request("v1/events/" + enc(String(id == null ? "" : id))),
      loopCurrent: () => request("v1/loop/current"),
      /* ---------- 既有只读面 ---------- */
      /* R1（§19.2）：scope 恒显式（cored 词形「写入/读取需显式 scope」——global 亦不省，
       * 省略=400「scope 与 storage 模板不相容」噪声 15 条/页+正文恒注水中）。 */
      object: (kind, scope, name) => request("v1/objects/" + enc(kind) + "/" + enc(name) + "?scope=" + enc(scope || "global")),
      kinds: () => request("v1/scheme/kinds"),
      drafts: () => request("v1/drafts"),
      draft: (kind, name, scope) => request("v1/drafts/" + enc(kind) + "/" + enc(name) + (scope && scope !== "global" ? "?scope=" + enc(scope) : "")),
      events: (limit = 200, params = {}) => {
        const q = new URLSearchParams({ limit: String(Number(limit) || 200) });
        Object.keys(params || {}).forEach((k) => {
          const v = params[k];
          if (v != null && v !== "") q.set(k, String(v));
        });
        return request("v1/events?" + q.toString());
      },
      workflows: () => request("v1/workflows"),
      workflowRuns: () => request("v1/workflows/runs"),
      tasks: () => request("v1/tasks"),
      taskSessions: (taskIds) => request("/core-console/task-sessions?taskIds=" + encodeURIComponent(taskIds.join(","))),
      sessions: () => request("v1/sessions"),
      session: (id) => request("v1/sessions/" + enc(String(id == null ? "" : id))),
      /* 会话上下文契约面（G-14——routes_sessions.go：ContextPack 词形投影+编辑分叉） */
      sessionContext: (id) => request("v1/sessions/" + enc(String(id == null ? "" : id)) + "/context"),
      saveSessionContext: (id, payload) => request("v1/sessions/" + enc(String(id == null ? "" : id)) + "/context", { method: "PUT", body: payload }),
      revertSessionContext: (id) => request("v1/sessions/" + enc(String(id == null ? "" : id)) + "/context/revert", { method: "POST" }),
      /* 三段式 Session 控制面（session.md——routes_session_config.go：GET/PUT config + 回整账只读） */
      sessionConfig: () => request("v1/session-config"),
      saveSessionConfig: (payload) => request("v1/session-config", { method: "PUT", body: payload }),
      sessionWritebacks: (n = 5) => request("v1/session-config/writebacks?n=" + Number(n)),
      inbox: () => request("v1/inbox"),
      proposals: () => request("v1/proposals"),
      approvals: (limit = 20) => request("v1/approvals?limit=" + Number(limit)),
      designDocs: () => request("v1/design-docs"),
      designDoc: (repo, path) => request("v1/design-docs/" + enc(repo) + "/" + enc(path)),
      /* V-16/V-17/V-18/V-21（2026-11-26 概念可见性批）：上下文索引 / 逐次正文史 / 历史快照 / 送达投影 / 代码审核只读面 */
      contextIndex: () => request("v1/context/index"),
      docHistory: (kind, scope, name) => request("v1/doc-history/" + enc(kind) + "/" + enc(name) + "?scope=" + enc(scope || "global")),
      revContent: (kind, scope, name, rv) => request("v1/rev-content/" + enc(kind) + "/" + enc(name) + "?scope=" + enc(scope || "global") + "&rv=" + enc(String(rv))),
      deliveries: () => request("v1/notifications/deliveries"),
      codeAuditFindings: () => request("v1/code-audit/findings"),
      /* ---------- 写通路（唯一入口——客户端零旁路） ---------- */
      apply: (payload) => request(P.apply, { method: "POST", body: payload }),
      saveDraft: (payload) => request(P.drafts, { method: "POST", body: payload }),
      /* 事件注入通路（§19 R3——人工批闸/点火/取消族；部署=宿主路由 /core-console/event-publish
       * （绝对词位——经 plugin 注入主体，source=human 服务端强制）。 */
      eventPub: (payload) => request("/core-console/event-pub", { method: "POST", body: payload })
    };
  }

  // src/core/router.js
  function parseHash(hash) {
    let h = String(hash || "").replace(/^#\/?/, "");
    h = h.split("?")[0];
    if (h === "") return [];
    return h.split("/").map((s) => {
      try {
        return decodeURIComponent(s);
      } catch {
        return s;
      }
    });
  }
  function matchSegs(pattern, segs) {
    const ps = String(pattern).split("/");
    if (ps.length !== segs.length) return null;
    const params = {};
    for (let i = 0; i < ps.length; i++) {
      const p = ps[i];
      const s = segs[i];
      if (p.charAt(0) === ":") params[p.slice(1)] = s;
      else if (p !== s) return null;
    }
    return params;
  }
  function createRouter(opts = {}) {
    const routes = [];
    let viewEl = null;
    let current = { route: "", pattern: "", segs: [], params: {} };
    function resolve(segs) {
      for (const r of routes) {
        const params = matchSegs(r.pattern, segs);
        if (params) return { r, params };
      }
      return null;
    }
    function notFoundHtml(segs) {
      return '<div class="view-inner"><div class="empty"><div class="e-title">\u672A\u77E5\u8DEF\u5F84\uFF08' + segs.map((s) => s).join("/") + '\uFF09</div><div class="small faint">\u8DEF\u7531\u8BCD\u5F62\u89C1\u63A7\u5236\u53F0\u5BFC\u822A\u2014\u2014\u6216\u4ECE\u9996\u9875\u5BF9\u8BDD\u8FDB\u5165\u3002</div><div class="e-act"><button class="btn btn-primary" data-go="home">\u56DE\u9996\u9875</button></div></div></div>';
    }
    const api2 = {
      /** 注册屏（pattern 如 'wf/:name/r/:n'；fn(viewEl, params, ctx)）。 */
      route(pattern, fn) {
        routes.push({ pattern: String(pattern), fn });
        return api2;
      },
      resolve,
      get current() {
        return current;
      },
      /** 导航（location.hash 变化触发 render；同址=强制重渲染）。 */
      go(path) {
        if (typeof location === "undefined") return;
        const h = "#/" + String(path || "").replace(/^\/+/, "");
        if (location.hash === h) api2.render();
        else location.hash = h;
      },
      /** 装配：挂容器 + 监听 hashchange + 首渲染。 */
      mount(el) {
        viewEl = el;
        if (typeof window !== "undefined") {
          window.addEventListener("hashchange", () => api2.render());
        }
        api2.render();
        return api2;
      },
      /** 解析当前 hash → 匹配屏执行（未命中=404 词形，不冒充内容）。 */
      render() {
        if (!viewEl || typeof location === "undefined") return null;
        const segs = parseHash(location.hash);
        if (!segs.length) segs.push(opts.defaultRoute || "home");
        const hit = resolve(segs);
        if (!hit) {
          current = { route: "404", pattern: "", segs, params: {} };
          viewEl.innerHTML = notFoundHtml(segs);
          return null;
        }
        current = { route: hit.r.pattern.split("/")[0], pattern: hit.r.pattern, segs, params: hit.params };
        hit.r.fn(viewEl, hit.params);
        return hit.r;
      }
    };
    return api2;
  }

  // src/core/ws.js
  function wsUrlOf(loc, path = "v1/stream") {
    const l = loc || (typeof location !== "undefined" ? location : null);
    if (!l || !l.host) return "/" + String(path).replace(/^\/+/, "");
    const proto = l.protocol === "https:" ? "wss://" : "ws://";
    return proto + l.host + "/" + String(path).replace(/^\/+/, "");
  }
  function hidden() {
    return typeof document !== "undefined" && !!document.hidden;
  }
  function connectStream(opts = {}) {
    const {
      url = wsUrlOf(),
      base = "",
      onSnapshot,
      onPatch,
      onState,
      pollMs = 3e4,
      maxFail = 5
    } = opts;
    const B = String(base || "").replace(/\/+$/, "");
    let ws = null;
    let state = "connecting";
    let fails = 0;
    let backoff = 1e3;
    let timer = null;
    let pollTimer = null;
    let pendingReconnect = false;
    let closed = false;
    function setState(s) {
      if (state === s) return;
      state = s;
      try {
        if (onState) onState(s);
      } catch {
      }
    }
    function clearTimers() {
      if (timer) {
        clearTimeout(timer);
        timer = null;
      }
      if (pollTimer) {
        clearInterval(pollTimer);
        pollTimer = null;
      }
    }
    function pollOnce() {
      if (closed || hidden()) return;
      fetch(B + "/v1/snapshot", { credentials: "same-origin" }).then((r) => r.ok ? r.json() : null).then((j) => {
        if (!j || closed) return;
        fails = 0;
        try {
          if (onSnapshot) onSnapshot(j);
        } catch {
        }
      }).catch(() => {
      });
    }
    function startPolling() {
      setState("degraded");
      if (pollTimer) return;
      pollOnce();
      pollTimer = setInterval(pollOnce, pollMs);
    }
    function scheduleReconnect() {
      if (closed) return;
      if (hidden()) {
        pendingReconnect = true;
        return;
      }
      if (timer) return;
      timer = setTimeout(() => {
        timer = null;
        open();
      }, backoff);
      backoff = Math.min(backoff * 2, 3e4);
    }
    function open() {
      if (closed) return;
      setState(fails > 0 ? "reconnecting" : "connecting");
      let sock;
      try {
        sock = new WebSocket(url);
      } catch {
        onFailure();
        return;
      }
      ws = sock;
      sock.onopen = () => {
        fails = 0;
        backoff = 1e3;
        if (pollTimer) {
          clearInterval(pollTimer);
          pollTimer = null;
        }
        setState("connected");
      };
      sock.onmessage = (ev) => {
        let msg = null;
        try {
          msg = JSON.parse(String(ev.data));
        } catch {
          return;
        }
        if (!msg || !msg.type) return;
        if (msg.type === "snapshot") {
          try {
            if (onSnapshot) onSnapshot(msg.snapshot !== void 0 ? msg.snapshot : msg.data !== void 0 ? msg.data : msg);
          } catch {
          }
        } else if (msg.type === "patch") {
          try {
            if (onPatch) onPatch(Array.isArray(msg.patch) ? msg.patch : [], msg);
          } catch {
          }
        } else if (msg.type === "ping") {
          try {
            if (sock.readyState === 1) sock.send(JSON.stringify({ type: "pong" }));
          } catch {
          }
        }
      };
      sock.onclose = () => {
        if (ws === sock) {
          ws = null;
          onFailure();
        }
      };
      sock.onerror = () => {
      };
    }
    function onFailure() {
      if (closed) return;
      fails += 1;
      if (fails >= maxFail) startPolling();
      else scheduleReconnect();
    }
    function resume() {
      if (closed) return;
      if (pendingReconnect) {
        pendingReconnect = false;
        scheduleReconnect();
        return;
      }
      if (ws) {
        if (state === "degraded") pollOnce();
        return;
      }
      if (state === "degraded") {
        backoff = 1e3;
        open();
        return;
      }
      if (state !== "connected") {
        open();
      }
    }
    function onVisibility() {
      if (hidden()) {
        if (timer) {
          clearTimeout(timer);
          timer = null;
          pendingReconnect = true;
        }
      } else {
        resume();
      }
    }
    if (typeof document !== "undefined") {
      document.addEventListener("visibilitychange", onVisibility);
    }
    open();
    return {
      close() {
        closed = true;
        clearTimers();
        if (typeof document !== "undefined") document.removeEventListener("visibilitychange", onVisibility);
        if (ws) {
          try {
            ws.close();
          } catch {
          }
          ws = null;
        }
      },
      get state() {
        return state;
      },
      send(obj) {
        if (ws && ws.readyState === 1) {
          try {
            ws.send(JSON.stringify(obj));
            return true;
          } catch {
            return false;
          }
        }
        return false;
      }
    };
  }

  // src/screens/home.js
  function faces() {
    return typeof window !== "undefined" && window.__CC && window.__CC.chatFaces || null;
  }
  var SAMPLES = [
    ["\u2713 \u4EE3\u7801\u5BA1\u6838", "\u8FDB\u884C\u4EE3\u7801\u5BA1\u6838\uFF0C\u770B\u770B\u5B9E\u9645\u548C\u6587\u6863\u5DEE\u8DDD\u6709\u591A\u5927\uFF0C\u5E76\u8FDB\u884C\u4FEE\u590D"],
    ["\u2315 \u627E\u6587\u6863", "\u627E\u51FA\u6240\u6709 drift \u7684\u6587\u6863\u5E76\u751F\u6210\u5BF9\u8D26\u6E05\u5355"],
    ["\u270E \u66F4\u65B0\u6587\u6863", "\u628A\u5F53\u524D\u5FAA\u73AF\u7684\u72B6\u6001\u5199\u6210\u4E00\u4EFD SessionLog \u6458\u8981\u5E76\u66F4\u65B0\u76F8\u5173\u6587\u6863"]
  ];
  function heroComposerHtml() {
    return '<div class="home-box" data-home-box><textarea class="home-text" data-home-text rows="2" placeholder="' + esc(t("\u63D0\u4EA4\u9700\u6C42\u3001\u627E\u6587\u6863\u3001\u66F4\u65B0\u6587\u6863\u2026\u2026")) + '" aria-label="' + esc(t("\u5BF9\u8BDD\u8F93\u5165")) + '"></textarea><div class="home-bar"><div class="home-tools">' + SAMPLES.map(([label, full]) => '<button type="button" class="home-tool" data-home-sample="' + esc(full) + '">' + esc(t(label)) + "</button>").join("") + '<label class="home-tool home-img" title="' + esc(t("\u9644\u52A0\u56FE\u7247\uFF08\u22648MB\uFF09")) + '">\u{1F5BC} ' + esc(t("\u56FE\u7247")) + '<input type="file" accept="image/png,image/jpeg,image/gif,image/webp,image/svg+xml" data-home-img hidden></label></div><button type="button" class="home-send" data-home-send aria-label="' + esc(t("\u53D1\u9001")) + '" title="' + esc(t("\u53D1\u9001")) + '"><span aria-hidden="true">\u2191</span></button></div></div><div class="home-note" data-home-note aria-live="polite"></div>';
  }
  function recentSessionsHtml(items) {
    const rows = (Array.isArray(items) ? items : []).filter((s) => s && s.sessionId).sort((a, b) => (Number(b.updatedAt) || 0) - (Number(a.updatedAt) || 0)).slice(0, 6);
    if (!rows.length) return "";
    return '<div class="home-recent-h">' + esc(t("\u6700\u8FD1\u4F1A\u8BDD")) + '</div><div class="home-recent-list">' + rows.map((s) => {
      const id = String(s.sessionId);
      const label = String(s.title || "").trim() || t("\u4F1A\u8BDD");
      return '<button type="button" class="home-recent-chip" data-go="' + esc(sessionRoute(id)) + '" title="' + esc(id) + '"><span class="home-recent-dot' + (s.running === true ? " live" : "") + '"></span><span class="home-recent-t">' + esc(trunc(label, 30)) + "</span>" + (s.updatedAt ? '<span class="home-recent-time">' + esc(relAge(Number(s.updatedAt)) || "") + "</span>" : "") + "</button>";
    }).join("") + '<button type="button" class="home-recent-chip all" data-go="chat/core-headless">' + esc(t("\u5168\u90E8\u4F1A\u8BDD")) + " \u2192</button></div>";
  }
  function sessionRoute(sessionId) {
    return "session/" + encodeURIComponent(String(sessionId));
  }
  function note(el, msg) {
    const n = el.querySelector("[data-home-note]");
    if (n) n.textContent = String(msg || "");
  }
  function autoGrow(ta) {
    if (!ta) return;
    ta.style.height = "auto";
    const max = 24 * 8 + 12;
    ta.style.height = Math.min(ta.scrollHeight, max) + "px";
  }
  async function submitHome(el, ctx2) {
    const ta = el.querySelector("[data-home-text]");
    const btn = el.querySelector("[data-home-send]");
    const text2 = String(ta && ta.value || "").trim();
    if (!text2) return false;
    const f = faces();
    if (!f || typeof f.send !== "function" || typeof f.create !== "function") {
      note(el, t("\u5BF9\u8BDD\u901A\u8DEF\u672A\u6CE8\u5165\uFF08chatfaces.js \u7F3A\u5E2D\uFF09\u2014\u2014\u663E\u5F0F\u7F3A\u5E2D\uFF0C\u4E0D\u5192\u5145\u5DF2\u53D1\u9001"));
      return false;
    }
    if (btn) btn.disabled = true;
    note(el, t("\u6B63\u5728\u521B\u5EFA\u4F1A\u8BDD\u2026"));
    let sid = "";
    let crStatus = 0;
    try {
      const cr = await f.create();
      crStatus = Number(cr && cr.status || 0);
      if (crStatus === 200 && cr.data && cr.data.sessionId) sid = String(cr.data.sessionId);
    } catch {
    }
    if (!sid) {
      if (btn) btn.disabled = false;
      note(el, t("\u65B0\u4F1A\u8BDD\u521B\u5EFA\u5931\u8D25") + "\uFF08" + (crStatus || "\u2014") + "\uFF09");
      return false;
    }
    note(el, t("\u6B63\u5728\u53D1\u9001\u2026"));
    let r;
    try {
      r = await f.send(sid, text2);
    } catch (e) {
      r = { status: 0, data: { error: String(e && e.message || e) } };
    }
    if (r && r.status === 200) {
      if (ta) ta.value = "";
      note(el, t("\u5DF2\u9001\u8FBE\uFF0C\u6B63\u5728\u6253\u5F00\u4F1A\u8BDD\u2026"));
      ctx2.go(sessionRoute(sid));
      return true;
    }
    if (btn) btn.disabled = false;
    note(el, t("\u53D1\u9001\u5931\u8D25") + "\uFF08" + String(r && r.status || 0) + (r && r.data && r.data.error ? " \xB7 " + String(r.data.error) : "") + "\uFF09");
    return false;
  }
  function bindImage(el) {
    const input = el.querySelector("[data-home-img]");
    const ta = el.querySelector("[data-home-text]");
    if (!input || !ta) return;
    input.addEventListener("change", async () => {
      const file = (input.files || [])[0];
      input.value = "";
      if (!file || !String(file.type || "").startsWith("image/")) return;
      if (file.size > 8 << 20) {
        note(el, t("\u56FE\u7247\u8D85\u8FC7 8MB \u4E0A\u9650"));
        return;
      }
      const f = faces();
      if (!f || typeof f.uploadImage !== "function") {
        note(el, t("\u56FE\u7247\u4E0A\u4F20\u901A\u8DEF\u672A\u6CE8\u5165"));
        return;
      }
      note(el, t("\u56FE\u7247\u4E0A\u4F20\u4E2D\u2026"));
      try {
        const url = await f.uploadImage(file);
        const tag2 = "![image](" + url + ")";
        ta.value = ta.value.trimEnd() + (ta.value.trimEnd() ? "\n" : "") + tag2;
        autoGrow(ta);
        note(el, t("\u56FE\u7247\u5DF2\u9644\u52A0\uFF08\u968F\u6D88\u606F\u53D1\u9001\uFF09"));
      } catch (e) {
        note(el, t("\u56FE\u7247\u4E0A\u4F20\u5931\u8D25") + "\uFF08" + String(e && e.message || e) + "\uFF09");
      }
    });
  }
  function bindHeroComposer(el, ctx2) {
    const ta = el.querySelector("[data-home-text]");
    const btn = el.querySelector("[data-home-send]");
    if (!ta || !btn) return;
    ta.addEventListener("input", () => autoGrow(ta));
    ta.addEventListener("keydown", (ev) => {
      if (ev.key === "Enter" && !ev.shiftKey) {
        ev.preventDefault();
        submitHome(el, ctx2);
      }
    });
    btn.addEventListener("click", () => submitHome(el, ctx2));
    el.querySelectorAll("[data-home-sample]").forEach((b) => {
      b.addEventListener("click", () => {
        const full = String(b.getAttribute("data-home-sample") || "");
        if (!full) return;
        ta.value = full;
        autoGrow(ta);
        ta.focus();
      });
    });
    bindImage(el);
    autoGrow(ta);
  }
  async function loadRecent(el) {
    const box = el.querySelector("[data-home-recent]");
    const f = faces();
    if (!box || !f || typeof f.sessions !== "function") return;
    try {
      const r = await f.sessions();
      if (!el.isConnected) return;
      const items = r && r.status === 200 && r.data && Array.isArray(r.data.items) ? r.data.items : [];
      box.innerHTML = recentSessionsHtml(items);
    } catch {
    }
  }
  function render(el, ctx2) {
    el.innerHTML = '<div class="home"><div class="home-hero"><h1 class="home-title">' + esc(t("\u6709\u4EC0\u4E48\u53EF\u4EE5\u5E2E\u5FD9\u7684\uFF1F")) + '</h1><div class="home-sub">' + esc(t("\u76F4\u63A5\u8BF4\u9700\u6C42\u2014\u2014\u627E\u6587\u6863\u3001\u66F4\u65B0\u6587\u6863\u3001\u63D0\u4EA4\u53D8\u66F4\uFF1B\u7ED3\u8BBA\u843D\u5230\u6587\u6863\u4E0E\u6D41\u7A0B\u3002")) + "</div>" + heroComposerHtml() + '<div class="home-recent" data-home-recent></div></div></div>';
    bindHeroComposer(el, ctx2);
    loadRecent(el);
  }

  // src/screens/modal.js
  function modalHost() {
    return typeof document !== "undefined" && document.querySelector(".cc-app") || (typeof document !== "undefined" ? document.body : null);
  }
  function confirmModal(opts) {
    const o = opts || {};
    return new Promise((resolve) => {
      const modal = document.createElement("div");
      modal.className = "chat-modal";
      modal.setAttribute("data-cc-modal", "1");
      const confirmCls = o.danger ? "btn btn-ghost btn-danger" : "btn btn-primary";
      modal.innerHTML = '<div class="chat-modal-mask" data-cc-modal-mask></div><div class="chat-modal-panel" style="width:420px;max-width:calc(100vw - 48px)"><div class="chat-modal-head"><div class="chat-modal-title">' + esc(t(o.title || t("\u786E\u8BA4\u64CD\u4F5C"))) + '</div></div><div class="chat-modal-body"><div class="small" style="line-height:1.6">' + (o.body || "") + '</div></div><div style="display:flex;justify-content:flex-end;gap:8px;padding:12px 16px 16px"><button type="button" class="btn btn-secondary sm" data-cc-modal-cancel>' + esc(t(o.cancelText || "\u53D6\u6D88")) + '</button><button type="button" class="' + confirmCls + ' sm" data-cc-modal-ok>' + esc(t(o.confirmText || "\u786E\u5B9A")) + "</button></div></div>";
      const done = (v) => {
        modal.remove();
        document.removeEventListener("keydown", onKey, true);
        resolve(v);
      };
      const onKey = (e) => {
        if (e.key === "Escape") done(false);
      };
      modal.querySelector("[data-cc-modal-ok]").addEventListener("click", () => done(true));
      modal.querySelector("[data-cc-modal-cancel]").addEventListener("click", () => done(false));
      modal.querySelector("[data-cc-modal-mask]").addEventListener("click", () => done(false));
      document.addEventListener("keydown", onKey, true);
      modalHost().appendChild(modal);
      const ok = modal.querySelector("[data-cc-modal-ok]");
      if (ok) ok.focus();
    });
  }
  function htmlModal(opts) {
    const o = opts || {};
    return new Promise((resolve) => {
      const modal = document.createElement("div");
      modal.className = "chat-modal";
      modal.setAttribute("data-cc-modal", "1");
      const w = o.wide ? "width:640px;max-width:calc(100vw - 48px)" : "width:420px;max-width:calc(100vw - 48px)";
      modal.innerHTML = '<div class="chat-modal-mask" data-cc-modal-mask></div><div class="chat-modal-panel" style="' + w + '"><div class="chat-modal-head"><div class="chat-modal-title">' + esc(t(o.title || t("\u8BE6\u60C5"))) + '</div><button type="button" class="btn btn-ghost sm" data-cc-modal-x aria-label="close">\u2715</button></div><div class="chat-modal-body" style="max-height:70vh;overflow:auto">' + (o.html || "") + "</div></div>";
      const done = () => {
        const body = modal.querySelector(".chat-modal-body");
        if (o.node && o.holder && body && o.node.parentElement === body) o.holder.appendChild(o.node);
        modal.remove();
        document.removeEventListener("keydown", onKey, true);
        resolve();
      };
      const onKey = (e) => {
        if (e.key === "Escape") done();
      };
      const x = modal.querySelector("[data-cc-modal-x]");
      if (x) x.addEventListener("click", done);
      modal.querySelector("[data-cc-modal-mask]").addEventListener("click", done);
      document.addEventListener("keydown", onKey, true);
      modalHost().appendChild(modal);
      if (o.node) {
        const body = modal.querySelector(".chat-modal-body");
        o.node.hidden = false;
        body.innerHTML = "";
        body.appendChild(o.node);
      }
    });
  }
  function toast(message, kind) {
    let wrap = document.querySelector(".chat-toast-wrap");
    if (!wrap) {
      wrap = document.createElement("div");
      wrap.className = "chat-toast-wrap";
      modalHost().appendChild(wrap);
    }
    const t2 = document.createElement("div");
    t2.className = "chat-toast" + (kind === "accent" ? " is-accent" : " is-error");
    t2.innerHTML = message;
    wrap.appendChild(t2);
    setTimeout(() => {
      t2.classList.add("is-leaving");
      setTimeout(() => t2.remove(), 260);
    }, 3500);
  }

  // src/render/chips.js
  function chip(label, cls = "mute", tiny = true, attrs = "") {
    return '<span class="chip ' + esc(cls) + (tiny ? " tiny" : "") + '"' + (attrs ? " " + attrs : "") + ">" + esc(label) + "</span>";
  }
  function chipGo(label, route, cls = "mute", tiny = true, title = "") {
    return '<span class="chip ' + esc(cls) + (tiny ? " tiny" : "") + ' click" data-go="' + esc(route) + '"' + (title ? ' title="' + esc(title) + '"' : "") + ">" + esc(label) + "</span>";
  }
  function kindBadge(kind) {
    return '<span class="chip kind tiny">' + esc(kind) + "</span>";
  }
  function kv(rows) {
    return '<div class="kv">' + (rows || []).filter((r) => r && r.length === 2).map(
      ([k, v]) => '<span class="k">' + esc(k) + '</span><span class="v">' + v + "</span>"
    ).join("") + "</div>";
  }
  function emptyBox(title, sub = "", actionHtml = "") {
    return '<div class="empty"><div class="e-title">' + esc(title) + "</div>" + (sub ? '<div class="small faint">' + esc(sub) + "</div>" : "") + (actionHtml ? '<div class="e-act">' + actionHtml + "</div>" : "") + "</div>";
  }
  function screenHead(title, sub = "", actionsHtml = "") {
    return '<div class="dh"><div class="dh-main"><div class="dh-title"><h1>' + esc(title) + "</h1></div>" + (sub ? '<div class="dh-sub">' + esc(sub) + "</div>" : "") + "</div>" + (actionsHtml ? '<div class="dh-actions">' + actionsHtml + "</div>" : "") + "</div>";
  }
  function encSeg(s) {
    return encodeURIComponent(String(s == null ? "" : s));
  }
  function objRoute(kind, scope, name) {
    return "obj/" + encSeg(kind) + "/" + encSeg(scope || "global") + "/" + encSeg(name);
  }
  function sessionTargetOf(sess) {
    const s = String(sess == null ? "" : sess).trim();
    if (!s) return "";
    if (/^session[-_]/i.test(s)) return "session/" + encSeg(s);
    if (s.includes("/")) return "task/" + encSeg(s);
    return "session/" + encSeg(s);
  }
  function tagRoute(tag2) {
    return "docs/tag/" + encSeg(tag2);
  }
  function docRoute(kind, scope, name) {
    return "doc/" + encSeg(kind) + "/" + encSeg(scope || "global") + "/" + encSeg(name);
  }
  function involvedRoute(s) {
    const r = String(s || "").trim();
    if (!r) return "";
    if (r.startsWith("#/")) {
      const segs = r.slice(2).split("/").filter(Boolean);
      return segs.length ? segs.map(encSeg).join("/") : "";
    }
    const m = /^([A-Za-z][A-Za-z0-9]*)\/([^/\s]+)\/([^/\s]+)$/.exec(r);
    if (m) return "obj/" + m.slice(1).map(encSeg).join("/");
    return "";
  }
  var DEFAULT_ORG = typeof window !== "undefined" && window.__CC && window.__CC.repoOrg || "friddle-self-recursive-agent";
  var REPO_NAMES = [
    "core_system",
    "core_design",
    "core_system_instance",
    "core_system_workservice",
    "core_system_dsh_plugin",
    "common_connector"
    // 部署侧私有仓名经 window.__CC.repoNames 注入（R-PUB：不入公开仓）
  ];
  function repoUrl(ref) {
    const r = String(ref || "").trim().replace(/^"+|"+$/g, "");
    if (!r) return "";
    if (/^https?:\/\//i.test(r)) return safeHref(r);
    const GH = (org, repo, sub, frag) => {
      const url = "https://github.com/" + org + "/" + String(repo || "").replace(/\.git$/, "");
      const s = (sub || "").replace(/^\/+/, "");
      return s ? url + "/blob/main/" + s + (frag ? "#" + frag : "") : url + (frag ? "#" + frag : "");
    };
    let m = /^(?:pkg|git|source):\/\/([^/\s]+)(?:\/([^\s#]*))?/.exec(r);
    if (m) {
      const first = m[1], rest = (m[2] || "").replace(/^\/+/, "");
      if (REPO_NAMES.includes(first)) return GH(DEFAULT_ORG, first, rest);
      return GH(first, rest.split("/")[0] || "", rest.split("/").slice(1).join("/"));
    }
    m = /^github\.com\/([^/\s]+)\/([^/\s#]+)(?:\/([^\s#]*))?(?:#(\S+))?/.exec(r);
    if (m) return GH(m[1], m[2], m[3] || "", m[4] || "");
    const bare = /^([A-Za-z0-9._-]+)(?:\/([^\s#]+))?$/.exec(r);
    if (bare && REPO_NAMES.includes(bare[1])) return GH(DEFAULT_ORG, bare[1], bare[2] || "");
    return "";
  }
  function repoLinkChip(ref, label) {
    const url = repoUrl(ref);
    if (!url) return "";
    const text2 = String(label || ref).trim() || ref;
    return '<a class="chip mute tiny click lk" href="' + esc(url) + '" target="_blank" rel="noopener noreferrer" title="' + esc(url) + '">' + esc(truncTxt(text2, 44)) + "</a>";
  }
  function truncTxt(s, n) {
    return s.length > n ? s.slice(0, n) + "\u2026" : s;
  }
  function syncBadge(runs, name) {
    const active = (runs || []).filter((r) => {
      const st = String(r && (r.status || r.state) || "");
      if (!st || st === "done" || st === "failed" || st === "canceled") return false;
      let hay = "";
      try {
        hay = JSON.stringify([r.involved, r.target, r.route, r.workflow]);
      } catch {
        hay = "";
      }
      return hay.includes(name);
    });
    if (!active.length) return chip("\u5DF2\u540C\u6B65", "ok");
    const top = active[0];
    const wf = String(top && (top.workflow || top.name) || "");
    const no = top && (top.run != null ? top.run : top.no != null ? top.no : "");
    if (wf && no !== "" && no != null) {
      return chipGo("\u540C\u6B65\u4E2D", "wf/" + encSeg(wf) + "/r/" + encSeg(no), "acc", true, "Workflow " + wf + " #" + no);
    }
    if (wf) return chipGo("\u540C\u6B65\u4E2D", "wf/" + encSeg(wf), "acc", true, "Workflow " + wf);
    return chip("\u540C\u6B65\u4E2D", "acc");
  }

  // ../../../web/node_modules/diff/libesm/diff/base.js
  var Diff = class {
    diff(oldStr, newStr, options2 = {}) {
      let callback;
      if (typeof options2 === "function") {
        callback = options2;
        options2 = {};
      } else if ("callback" in options2) {
        callback = options2.callback;
      }
      const oldString = this.castInput(oldStr, options2);
      const newString = this.castInput(newStr, options2);
      const oldTokens = this.removeEmpty(this.tokenize(oldString, options2));
      const newTokens = this.removeEmpty(this.tokenize(newString, options2));
      return this.diffWithOptionsObj(oldTokens, newTokens, options2, callback);
    }
    diffWithOptionsObj(oldTokens, newTokens, options2, callback) {
      var _a2;
      const done = (value) => {
        value = this.postProcess(value, options2);
        if (callback) {
          setTimeout(function() {
            callback(value);
          }, 0);
          return void 0;
        } else {
          return value;
        }
      };
      const newLen = newTokens.length, oldLen = oldTokens.length;
      let editLength = 1;
      let maxEditLength = newLen + oldLen;
      if (options2.maxEditLength != null) {
        maxEditLength = Math.min(maxEditLength, options2.maxEditLength);
      }
      const maxExecutionTime = (_a2 = options2.timeout) !== null && _a2 !== void 0 ? _a2 : Infinity;
      const abortAfterTimestamp = Date.now() + maxExecutionTime;
      const bestPath = [{ oldPos: -1, lastComponent: void 0 }];
      let newPos = this.extractCommon(bestPath[0], newTokens, oldTokens, 0, options2);
      if (bestPath[0].oldPos + 1 >= oldLen && newPos + 1 >= newLen) {
        return done(this.buildValues(bestPath[0].lastComponent, newTokens, oldTokens));
      }
      let minDiagonalToConsider = -Infinity, maxDiagonalToConsider = Infinity;
      const execEditLength = () => {
        for (let diagonalPath = Math.max(minDiagonalToConsider, -editLength); diagonalPath <= Math.min(maxDiagonalToConsider, editLength); diagonalPath += 2) {
          let basePath;
          const removePath = bestPath[diagonalPath - 1], addPath = bestPath[diagonalPath + 1];
          if (removePath) {
            bestPath[diagonalPath - 1] = void 0;
          }
          let canAdd = false;
          if (addPath) {
            const addPathNewPos = addPath.oldPos - diagonalPath;
            canAdd = addPath && 0 <= addPathNewPos && addPathNewPos < newLen;
          }
          const canRemove = removePath && removePath.oldPos + 1 < oldLen;
          if (!canAdd && !canRemove) {
            bestPath[diagonalPath] = void 0;
            continue;
          }
          if (!canRemove || canAdd && removePath.oldPos < addPath.oldPos) {
            basePath = this.addToPath(addPath, true, false, 0, options2);
          } else {
            basePath = this.addToPath(removePath, false, true, 1, options2);
          }
          newPos = this.extractCommon(basePath, newTokens, oldTokens, diagonalPath, options2);
          if (basePath.oldPos + 1 >= oldLen && newPos + 1 >= newLen) {
            return done(this.buildValues(basePath.lastComponent, newTokens, oldTokens)) || true;
          } else {
            bestPath[diagonalPath] = basePath;
            if (basePath.oldPos + 1 >= oldLen) {
              maxDiagonalToConsider = Math.min(maxDiagonalToConsider, diagonalPath - 1);
            }
            if (newPos + 1 >= newLen) {
              minDiagonalToConsider = Math.max(minDiagonalToConsider, diagonalPath + 1);
            }
          }
        }
        editLength++;
      };
      if (callback) {
        (function exec() {
          setTimeout(function() {
            if (editLength > maxEditLength || Date.now() > abortAfterTimestamp) {
              return callback(void 0);
            }
            if (!execEditLength()) {
              exec();
            }
          }, 0);
        })();
      } else {
        while (editLength <= maxEditLength && Date.now() <= abortAfterTimestamp) {
          const ret = execEditLength();
          if (ret) {
            return ret;
          }
        }
      }
    }
    addToPath(path, added, removed, oldPosInc, options2) {
      const last = path.lastComponent;
      if (last && !options2.oneChangePerToken && last.added === added && last.removed === removed) {
        return {
          oldPos: path.oldPos + oldPosInc,
          lastComponent: { count: last.count + 1, added, removed, previousComponent: last.previousComponent }
        };
      } else {
        return {
          oldPos: path.oldPos + oldPosInc,
          lastComponent: { count: 1, added, removed, previousComponent: last }
        };
      }
    }
    extractCommon(basePath, newTokens, oldTokens, diagonalPath, options2) {
      const newLen = newTokens.length, oldLen = oldTokens.length;
      let oldPos = basePath.oldPos, newPos = oldPos - diagonalPath, commonCount = 0;
      while (newPos + 1 < newLen && oldPos + 1 < oldLen && this.equals(oldTokens[oldPos + 1], newTokens[newPos + 1], options2)) {
        newPos++;
        oldPos++;
        commonCount++;
        if (options2.oneChangePerToken) {
          basePath.lastComponent = { count: 1, previousComponent: basePath.lastComponent, added: false, removed: false };
        }
      }
      if (commonCount && !options2.oneChangePerToken) {
        basePath.lastComponent = { count: commonCount, previousComponent: basePath.lastComponent, added: false, removed: false };
      }
      basePath.oldPos = oldPos;
      return newPos;
    }
    equals(left, right, options2) {
      if (options2.comparator) {
        return options2.comparator(left, right);
      } else {
        return left === right || !!options2.ignoreCase && left.toLowerCase() === right.toLowerCase();
      }
    }
    removeEmpty(array) {
      const ret = [];
      for (let i = 0; i < array.length; i++) {
        if (array[i]) {
          ret.push(array[i]);
        }
      }
      return ret;
    }
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    castInput(value, options2) {
      return value;
    }
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    tokenize(value, options2) {
      return Array.from(value);
    }
    join(chars) {
      return chars.join("");
    }
    postProcess(changeObjects, options2) {
      return changeObjects;
    }
    get useLongestToken() {
      return false;
    }
    buildValues(lastComponent, newTokens, oldTokens) {
      const components = [];
      let nextComponent;
      while (lastComponent) {
        components.push(lastComponent);
        nextComponent = lastComponent.previousComponent;
        delete lastComponent.previousComponent;
        lastComponent = nextComponent;
      }
      components.reverse();
      const componentLen = components.length;
      let componentPos = 0, newPos = 0, oldPos = 0;
      for (; componentPos < componentLen; componentPos++) {
        const component = components[componentPos];
        if (!component.removed) {
          if (!component.added && this.useLongestToken) {
            let value = newTokens.slice(newPos, newPos + component.count);
            value = value.map(function(value2, i) {
              const oldValue = oldTokens[oldPos + i];
              return oldValue.length > value2.length ? oldValue : value2;
            });
            component.value = this.join(value);
          } else {
            component.value = this.join(newTokens.slice(newPos, newPos + component.count));
          }
          newPos += component.count;
          if (!component.added) {
            oldPos += component.count;
          }
        } else {
          component.value = this.join(oldTokens.slice(oldPos, oldPos + component.count));
          oldPos += component.count;
        }
      }
      return components;
    }
  };

  // ../../../web/node_modules/diff/libesm/diff/line.js
  var LineDiff = class extends Diff {
    constructor() {
      super(...arguments);
      this.tokenize = tokenize;
    }
    equals(left, right, options2) {
      if (options2.ignoreWhitespace) {
        if (!options2.newlineIsToken || !left.includes("\n")) {
          left = left.trim();
        }
        if (!options2.newlineIsToken || !right.includes("\n")) {
          right = right.trim();
        }
      } else if (options2.ignoreNewlineAtEof && !options2.newlineIsToken) {
        if (left.endsWith("\n")) {
          left = left.slice(0, -1);
        }
        if (right.endsWith("\n")) {
          right = right.slice(0, -1);
        }
      }
      return super.equals(left, right, options2);
    }
  };
  var lineDiff = new LineDiff();
  function diffLines(oldStr, newStr, options2) {
    return lineDiff.diff(oldStr, newStr, options2);
  }
  function tokenize(value, options2) {
    if (options2.stripTrailingCr) {
      value = value.replace(/\r\n/g, "\n");
    }
    const retLines = [], linesAndNewlines = value.split(/(\n|\r\n)/);
    if (!linesAndNewlines[linesAndNewlines.length - 1]) {
      linesAndNewlines.pop();
    }
    for (let i = 0; i < linesAndNewlines.length; i++) {
      const line = linesAndNewlines[i];
      if (i % 2 && !options2.newlineIsToken) {
        retLines[retLines.length - 1] += line;
      } else {
        retLines.push(line);
      }
    }
    return retLines;
  }

  // src/render/diff.js
  function diffLines2(oldStr, newStr) {
    const parts = diffLines(String(oldStr == null ? "" : oldStr), String(newStr == null ? "" : newStr));
    const rows = [];
    for (const p of parts) {
      const lines = String(p.value == null ? "" : p.value).split("\n");
      if (lines.length && lines[lines.length - 1] === "") lines.pop();
      const t2 = p.added ? "+" : p.removed ? "-" : " ";
      for (const s of lines) rows.push({ t: t2, s });
    }
    return compress(rows);
  }
  function compress(rows) {
    const keep = new Array(rows.length).fill(false);
    rows.forEach((r, i) => {
      if (r.t !== " ") {
        for (let j = Math.max(0, i - 2); j <= Math.min(rows.length - 1, i + 2); j++) keep[j] = true;
      }
    });
    const out = [];
    let last = -3;
    rows.forEach((r, i) => {
      if (!keep[i]) return;
      if (i - last > 1 && r.t === " ") out.push({ t: "gap", s: "\u22EF" });
      out.push(r);
      last = i;
    });
    if (!out.length) out.push({ t: "ctx", s: t("\uFF08\u65E0\u5DEE\u5F02\uFF09") });
    return out;
  }
  function renderDiff(lines, maxLines = 200) {
    const cap = Number(maxLines) || 200;
    const html3 = lines.slice(0, cap).map((l) => {
      if (l.t === "gap") return '<div class="diff-line ctx">   \u22EF</div>';
      const sign = l.t === "+" ? "+" : l.t === "-" ? "-" : " ";
      const cls = l.t === "+" ? "add" : l.t === "-" ? "del" : "ctx";
      return '<div class="diff-line ' + cls + '"><span class="d-sign">' + sign + "</span>" + esc(trunc(l.s, 500)) + "</div>";
    }).join("");
    const tail = lines.length > cap ? t('<div class="diff-line ctx">\u2026 \u5171 ') + lines.length + t(" \u884C\u5DEE\u5F02</div>") : "";
    return '<div class="diff">' + html3 + tail + "</div>";
  }
  function diffStats(lines) {
    let add = 0;
    let del = 0;
    (lines || []).forEach((l) => {
      if (l.t === "+") add += 1;
      else if (l.t === "-") del += 1;
    });
    return { add, del };
  }
  function hunksOf(lines) {
    let n = 0;
    let inHunk = false;
    (lines || []).forEach((l) => {
      if (l.t === "+" || l.t === "-") {
        if (!inHunk) {
          n += 1;
          inHunk = true;
        }
      } else {
        inHunk = false;
      }
    });
    return n;
  }

  // src/screens/approvals.js
  var ignored = /* @__PURE__ */ new Set();
  function pendingItems(ctx2) {
    const drafts = ctx2.store.drafts().filter((d) => d && d.kind && d.name).map((d) => ({ kind: "draft", key: "draft|" + d.kind + "|" + (d.scope || "global") + "|" + d.name, draft: d }));
    const runs = ctx2.store.runs().filter((r) => r && (String(r.status || r.state || "") === "suspended" || String(r.phase || "") === "AwaitingApproval")).map((r) => ({ kind: "run", key: "run|" + String(r.workflow || "") + "#" + String(r.run != null ? r.run : ""), run: r }));
    const props = ctx2.store.proposals().filter((p) => p && p.name && String(p.status || "submitted") === "submitted").map((p) => ({ kind: "proposal", key: "proposal|" + String(p.scope || "global") + "|" + String(p.name), proposal: p }));
    const evs = ctx2.store.events().filter((e) => e && e.type === "ApprovalNeeded" && !ignored.has("ev|" + String(e.id ?? ""))).map((e) => ({ kind: "event", key: "ev|" + String(e.id ?? ""), ev: e }));
    return drafts.filter((x) => !ignored.has(x.key)).concat(runs, props, evs);
  }
  function apKeyOf(d) {
    return (d.kind + "/" + (d.scope || "global") + "/" + d.name).replace(/[^A-Za-z0-9]/g, "_");
  }
  function runApprovalInfoOf(store2, wf, scope, run, startedAt) {
    const anchor = "Workflow/" + String(scope || "global") + "/" + String(wf || "");
    const ev = (store2.events() || []).find((e) => e && e.type === "ApprovalNeeded" && String(e.involved || "").indexOf(anchor) >= 0);
    const docs = (store2.events() || []).filter((e) => e && e.type === "DocApplied").filter((e) => !startedAt || String(e.ts || e.at || "") >= String(startedAt)).slice(0, 5).map((e) => String(e.involved || "").split(",")[0].trim()).filter(Boolean);
    const drs = (store2.drafts() || []).filter((d) => d && d.kind && d.name).slice(0, 5).map((d) => d.kind + "/" + (d.scope || "global") + "/" + d.name);
    return { judge: ev ? String(ev.summary || "") : "", docs, drafts: drs };
  }
  function runApprovalInfoHtml(info) {
    const parts = [];
    if (info.docs.length) {
      parts.push('<span class="chip ok tiny">\u5DF2\u66F4\u65B0 ' + info.docs.length + ' \u7BC7</span> <span class="mono small">' + esc(info.docs.map((d) => trunc(d, 30)).join("\u3001")) + "</span>");
    }
    if (info.drafts.length) {
      parts.push('<span class="chip warn tiny">\u5F85\u6279\u7A3F ' + info.drafts.length + ' \u4EFD</span> <span class="mono small">' + esc(info.drafts.map((d) => trunc(d, 30)).join("\u3001")) + "</span>");
    }
    if (info.judge) {
      parts.push('<span class="small">' + esc(t("\u5224\u65AD\uFF1A")) + esc(trunc(info.judge, 110)) + "</span>");
    }
    if (!parts.length) return '<span class="small faint">' + esc(t("\u7A97\u53E3\u5185\u65E0\u6587\u6863\u52A8\u6001\uFF08DocApplied/\u5F85\u6279\u7A3F\u7F3A\u5E2D\uFF09\u2014\u2014\u6279\u51C6\u524D\u8BE6\u60C5\u8BF7\u770B\u300C\u8FD0\u884C\u8BE6\u60C5\u300D\u3002")) + "</span>";
    return parts.join(' <span class="faint">\xB7</span> ');
  }
  function wfAnchorOf(ev) {
    const first = String(ev && ev.involved || "").split(",")[0].trim();
    return /^Workflow\/[^/\s]+\/[^/\s]+$/.test(first) ? first : "";
  }
  function matchesApprovalFocus(it, focus) {
    const f = String(focus || "").trim();
    if (!f || !it) return false;
    if (it.kind === "draft" && it.draft) {
      const d = it.draft;
      const key = d.kind + "/" + (d.scope || "global") + "/" + d.name;
      return apKeyOf(d) === f || key === f || String(d.name) === f;
    }
    if (it.kind === "run" && it.run) {
      const r = it.run;
      return String(r.workflow || "") === f || "Workflow/" + String(r.owner || "global") + "/" + String(r.workflow || "") === f || String(r.run) === f;
    }
    if (it.kind === "proposal" && it.proposal) {
      const p = it.proposal;
      return String(p.name) === f || "Proposal/" + String(p.scope || "global") + "/" + String(p.name) === f;
    }
    if (it.kind === "event" && it.ev) {
      const e = it.ev;
      return String(e.involved || "").includes(f) || String(e.summary || "").includes(f) || String(e.dedupKey || "").includes(f);
    }
    return false;
  }
  function itemHtml(ctx2, it, idx, focused = false) {
    const approveCls = idx === 0 ? "btn btn-primary sm" : "btn btn-secondary sm";
    const hl = focused ? ' data-ap-focused="1" style="outline:2px solid var(--accent)"' : "";
    if (it.kind === "run") {
      const r = it.run;
      const a0 = verdictAnchorOf(r);
      const anchor = "Workflow/" + String(r.owner || "global") + "/" + String(r.workflow || "");
      const no = String(r.run != null ? r.run : "");
      const info = runApprovalInfoHtml(runApprovalInfoOf(ctx2.store, String(r.workflow || ""), String(r.owner || "global"), no, String(r.startedAt || "")));
      return '<div class="card ap-item"' + hl + '><div class="card-h"><span class="h-title mono">' + esc(String(r.workflow || "\u2014")) + (no ? " #" + esc(no) : "") + '</span><span class="h-right">' + chip(String(r.phase || "AwaitingApproval"), "warn") + '</span></div><div class="card-b"><div class="small">' + esc(trunc(String(r.involved || r.route || t("\u5DE5\u4F5C\u6D41\u95F8\u951A\u6302\u8D77\u2014\u2014\u7B49\u5F85\u5BA1\u6279\u4FE1\u53F7")), 160)) + '</div><div class="small" style="margin-top:8px;border-top:1px dashed var(--border);padding-top:8px">' + info + '</div><div data-apverdict="' + esc(apVerdictKeyOf(r, a0)) + '"><div class="small faint" style="margin-top:8px">' + esc(t("\u5224\u51B3\u6587\u6863\uFF08ops_core_inner_agent\uFF09\u73B0\u8BFB\u4E2D\u2026")) + '</div></div><div style="margin-top:8px;display:flex;gap:6px;flex-wrap:wrap">' + chipGo(t("\u951A ") + anchor, "wf/" + encodeURIComponent(String(r.workflow || "")) + (no ? "/r/" + encodeURIComponent(no) : ""), "acc", true, t("\u8FD0\u884C\u8BE6\u60C5\u2014\u2014\u6279\u95F8\u951A")) + '</div><div class="dh-actions" style="margin-top:10px"><button class="' + approveCls + '" data-ap-grant="' + esc(it.key) + '" data-ap-anchor="' + esc(anchor) + '" data-ap-run="' + esc(no) + '">' + esc(t("\u6279\u51C6\u7EE7\u7EED")) + '</button> <button class="btn btn-ghost btn-danger sm" data-ap-reject="' + esc(it.key) + '" data-ap-anchor="' + esc(anchor) + '" data-ap-run="' + esc(no) + '">' + esc(t("\u9A73\u56DE\u7EC8\u6B62")) + "</button></div></div></div>";
    }
    if (it.kind === "proposal") {
      const p = it.proposal;
      const anchor = "Proposal/" + String(p.scope || "global") + "/" + String(p.name);
      return '<div class="card ap-item"' + hl + '><div class="card-h"><span class="h-title">' + esc(String(p.title || p.name || t("\u672A\u547D\u540D\u63D0\u6848"))) + '</span><span class="h-right">' + chip(String(p.status || "submitted"), "warn") + '</span></div><div class="card-b"><div class="small">' + esc(trunc(String(p.summary || ""), 200) || t("\uFF08summary \u7F3A\u5E2D\u2014\u2014\u8BE6\u60C5\u89C1\u5BF9\u8C61\u5DE5\u4F5C\u53F0\uFF09")) + '</div><div style="margin-top:8px;display:flex;gap:6px;flex-wrap:wrap">' + chipGo(t("\u951A ") + anchor, "obj/Proposal/" + encodeURIComponent(String(p.scope || "global")) + "/" + encodeURIComponent(String(p.name)), "mute", true, t("\u5BF9\u8C61\u5DE5\u4F5C\u53F0")) + '</div><div class="dh-actions" style="margin-top:10px"><button class="' + approveCls + '" data-ap-prop="' + esc(it.key) + '" data-ap-prop-anchor="' + esc(anchor) + '" data-ap-prop-decision="approve">' + esc(t("\u6279\u51C6\uFF08= Apply\uFF09")) + '</button> <button class="btn btn-ghost btn-danger sm" data-ap-prop="' + esc(it.key) + '" data-ap-prop-anchor="' + esc(anchor) + '" data-ap-prop-decision="reject">' + esc(t("\u9A73\u56DE")) + "</button></div></div></div>";
    }
    if (it.kind === "event") {
      const e = it.ev;
      const involved = String(e.involved || "");
      const anchor = wfAnchorOf(e);
      const wfName = anchor ? String(anchor.split("/")[2] || "") : "";
      const route = involvedRoute(involved.split(",")[0] || "");
      const gateBtns = anchor ? '<button class="' + approveCls + '" data-ap-grant="' + esc(it.key) + '" data-ap-ev="' + esc(String(e.id == null ? "" : e.id)) + '" data-ap-anchor="' + esc(anchor) + '">' + esc(t("\u6279\u51C6\u7EE7\u7EED")) + '</button> <button class="btn btn-ghost btn-danger sm" data-ap-reject="' + esc(it.key) + '" data-ap-ev="' + esc(String(e.id == null ? "" : e.id)) + '" data-ap-anchor="' + esc(anchor) + '">' + esc(t("\u9A73\u56DE\u7EC8\u6B62")) + "</button> " : '<button class="btn btn-secondary sm" disabled title="' + esc(t("\u4E8B\u4EF6\u672A\u643A\u5E26 Workflow \u951A")) + '">' + esc(t("\u6279\u51C6\u7EE7\u7EED")) + '</button> <button class="btn btn-ghost btn-danger sm" disabled title="' + esc(t("\u4E8B\u4EF6\u672A\u643A\u5E26 Workflow \u951A")) + '">' + esc(t("\u9A73\u56DE\u7EC8\u6B62")) + "</button> ";
      return '<div class="card ap-item"' + hl + '><div class="card-h"><span class="h-title">' + esc(String(e.type)) + '</span><span class="h-right">' + chip(String(e.level || "warn"), "warn") + '</span></div><div class="card-b"><div class="small">' + esc(trunc(String(e.summary || e.involved || ""), 160)) + '</div><div style="margin-top:8px;display:flex;gap:6px;flex-wrap:wrap">' + (route ? chipGo(involved.split(",")[0], route, "mute", true, t("\u53BB\u5BF9\u8C61\u5DE5\u4F5C\u53F0")) : chip(involved || "\u2014", "mute")) + (wfName ? chipGo(t("\u951A ") + anchor, "wf/" + encodeURIComponent(wfName), "acc", true, t("Workflow \u8BE6\u60C5\u2014\u2014\u6279\u95F8\u951A")) : "") + '</div><div class="dh-actions" style="margin-top:10px">' + gateBtns + '<button class="btn btn-ghost sm" data-ap-ignore="' + esc(it.key) + '">' + esc(t("\u5FFD\u7565")) + "</button></div></div></div>";
    }
    const d = it.draft;
    const scope = d.scope || "global";
    return '<div class="card ap-item" data-ap-item="' + esc(apKeyOf(d)) + '"' + hl + '><div class="card-h"><span class="h-title"><span class="lk mono" data-go="' + esc(docRoute(d.kind, scope, d.name)) + '">' + esc(scope + "/" + d.name) + '</span></span><span class="h-right">' + kindBadge(d.kind) + " " + chip(t("\u7A3F: ") + (d.fieldManager || "\u2014"), "mute") + '</span></div><div class="card-b"><div data-apdiff="' + esc(apKeyOf(d)) + t('"><div class="small faint">diff \u73B0\u8BFB\u4E2D\u2026</div></div>') + '<div class="dh-actions" style="margin-top:10px"><button class="' + approveCls + '" data-ap-approve="' + esc(apKeyOf(d)) + t('">\u6279\u51C6\uFF08= Apply\uFF09</button> ') + '<button class="btn btn-ghost sm" data-ap-ignore="' + esc(it.key) + t('">\u5FFD\u7565</button> ') + '<button class="btn btn-secondary sm" data-go="' + esc(docRoute(d.kind, scope, d.name)) + t('">\u7F16\u8F91\u5668</button>') + "</div></div></div>";
  }
  async function hydrateDiff(ctx2, d) {
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const [draftR, faceR] = await Promise.all([
      ctx2.api.draft(d.kind, d.name, d.scope || "global"),
      ctx2.api.object(d.kind, d.scope || "global", d.name)
    ]);
    if ((typeof location !== "undefined" ? location.hash : "") !== stamp) return;
    const box = document.querySelector('[data-apdiff="' + apKeyOf(d) + '"]');
    if (!box) return;
    if (!draftR.ok) {
      box.innerHTML = t('<div class="small faint">\u6B63\u6587\u6682\u4E0D\u53EF\u8BFB\uFF08') + esc(String(draftR.error || draftR.status)) + t("\uFF09\u2014\u2014\u91CD\u8FDB\u5373\u6062\u590D\u3002</div>");
      hydrateVerdict(ctx2, d);
      return;
    }
    const dr = draftR.data && typeof draftR.data === "object" ? draftR.data : {};
    const working = String(dr.content || "");
    let base = t("\uFF08\u5BF9\u8C61\u5C1A\u672A\u843D\u5E93\u2014\u2014\u6279\u51C6\u5373\u521B\u5EFA\uFF09");
    let baseNote = t("\u57FA\u51C6=r?");
    if (faceR.ok) {
      const data = faceR.data && typeof faceR.data === "object" ? faceR.data : {};
      base = data.content != null ? String(data.content) : "";
      baseNote = t("\u57FA\u51C6=r") + (dr.baseResourceVersion != null ? dr.baseResourceVersion : data.resourceVersion != null ? data.resourceVersion : "?");
    } else if (faceR.status !== 404) {
      base = t("\uFF08\u9762\u73B0\u503C\u6682\u4E0D\u53EF\u8BFB\uFF1A") + String(faceR.error || faceR.status) + "\uFF09";
    }
    const stats = diffStats(diffLines2(base, working));
    box.innerHTML = renderDiff(diffLines2(base, working), 26) + '<div class="small faint" style="margin-top:4px">' + esc(baseNote) + t(" \xB7 \u7A3F ") + working.length + t("B vs \u843D\u5E93\u503C ") + base.length + t("B\uFF08\u73B0\u8BFB\uFF0C\u4E0D\u7F13\u5B58\uFF09 \xB7 ") + '<span class="chip ok tiny">+' + stats.add + '</span> <span class="chip esc tiny">-' + stats.del + '</span></div><div data-apverdict="' + esc(apKeyOf(d)) + '"><div class="small faint">' + esc(t("\u5224\u51B3\u6587\u6863\uFF08ops_core_inner_agent\uFF09\u73B0\u8BFB\u4E2D\u2026")) + "</div></div>";
    hydrateVerdict(ctx2, d);
  }
  function verdictDocNames(d) {
    const base = "docdiff-" + String(d.kind + "-" + d.name).toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "");
    return d.run != null && d.run !== "" && Number(d.run) > 0 ? [base + "-run" + Number(d.run), base] : [base];
  }
  function verdictAnchorOf(r) {
    const inv = String(r && r.involved || "").split(",")[0].trim();
    const m = /^(Workflow)\/([^/\s]+)\/([^/\s]+)$/.exec(inv);
    if (m) return { kind: m[1], scope: m[2], name: m[3] };
    return { kind: "Workflow", scope: String(r && r.owner || "global"), name: String(r && r.workflow || "") };
  }
  function apVerdictKeyOf(r, anchor) {
    const base = apKeyOf(anchor || verdictAnchorOf(r));
    const no = String(r && r.run != null ? r.run : "");
    return no ? base + "-run" + no : base;
  }
  async function hydrateVerdict(ctx2, d) {
    try {
      const candidates = verdictDocNames(d);
      let hit = null;
      for (const nm of candidates) {
        const r = await ctx2.api.object("DocumentDiff", "global", nm);
        if (r && r.ok && r.data && r.data.content) {
          hit = { name: nm, content: String(r.data.content) };
          break;
        }
      }
      const vkey = d.apVerdictKey || apKeyOf(d);
      const vbox = document.querySelector('[data-apverdict="' + CSS.escape(vkey) + '"]');
      if (!vbox) return;
      if (hit) {
        const raw = hit.content;
        const vm = raw.match(/verdict:\s*\|\n([\s\S]*?)(?:\n  [a-zA-Z]|\n\w|$)/);
        const cm = raw.match(/changed:\s*"?([^"\n]+)"?/);
        const verdict = vm ? vm[1].trim() : raw;
        vbox.innerHTML = '<div class="card" style="border:1px solid var(--ds-accent);margin-top:8px"><div class="card-h"><span class="h-title">' + esc(t("Diff \u5224\u51B3 \xB7 ops_core_inner_agent")) + '</span><span class="h-right">' + esc(cm ? cm[1] : "") + '</span></div><div class="card-b"><div class="small" style="white-space:pre-wrap">' + esc(verdict) + '</div><div class="small faint" style="margin-top:6px">' + esc(t("\u5BF9\u8C61 DocumentDiff/global/")) + esc(hit.name) + "</div></div></div>";
      } else {
        vbox.innerHTML = '<div class="small faint" style="margin-top:8px">' + esc(t("\u5224\u51B3\u6587\u6863\u5C1A\u672A\u751F\u6210\uFF08DocumentDiff/global/")) + esc(candidates.join("\u3001")) + esc(t(" \u7F3A\u5E2D\u2014\u2014\u5F15\u64CE\u4EBA\u5BA1\u6302\u8D77\u5373\u81EA\u52A8\u751F\u6210\uFF1B\u7F3A\u5E2D=\u672A\u5230\u751F\u6210\u7A97\u53E3\uFF0C\u4E0D\u7F16\u9020")) + "</div>";
      }
    } catch (e) {
    }
  }
  async function approve(ctx2, d, btn) {
    btn.disabled = true;
    btn.textContent = t("\u9884\u68C0\u4E2D\u2026");
    const dry = await ctx2.api.apply({
      kind: d.kind,
      scope: d.scope || "global",
      name: d.name,
      fromDraft: true,
      dryRun: true,
      fieldManager: "console:approve",
      reasonRef: "console:approve-dry"
    });
    btn.disabled = false;
    btn.textContent = t("\u6279\u51C6\uFF08= Apply\uFF09");
    if (!dry.ok) {
      toast(esc(t("dry-run \u9884\u68C0\u672A\u8FC7\uFF08") + String(dry.error || dry.status) + t("\uFF09\u2014\u2014\u96F6\u526F\u4F5C\u7528\uFF0C\u7A3F\u5DF2\u4FDD\u7559\u3002")), "error");
      return;
    }
    const cur = dry.data && dry.data.rv != null ? Number(dry.data.rv) : null;
    const expect = d.baseResourceVersion != null ? Number(d.baseResourceVersion) : null;
    const lockNote = cur != null && expect != null ? cur === expect ? t("\u7248\u672C\u5BF9\u9F50 r") + cur : t("\u7248\u672C\u5DF2\u53D8\u5316\uFF08\u73B0 r") + cur + t(" vs \u7F16\u8F91\u57FA\u7EBF r") + expect + t("\u2014\u2014\u843D\u5E93\u524D\u8BF7\u5148\u5237\u65B0\uFF09") : "";
    const approved = await confirmModal({
      title: t("\u6279\u51C6\uFF08= Apply\uFF09"),
      body: esc(t("\u786E\u8BA4\u6279\u51C6\uFF08Apply\uFF09") + d.kind + "/" + d.name + t("\uFF1Fdry-run \u5DF2\u8FC7\uFF08rv=") + String(cur) + "\uFF09" + (lockNote ? "\uFF1B" + lockNote : "") + "\u3002"),
      confirmText: t("\u6279\u51C6")
    });
    if (!approved) return;
    btn.disabled = true;
    btn.textContent = t("\u843D\u5E93\u4E2D\u2026");
    const r = await ctx2.api.apply({
      kind: d.kind,
      scope: d.scope || "global",
      name: d.name,
      fromDraft: true,
      fieldManager: "console:approve",
      reasonRef: "console:approve"
    });
    btn.disabled = false;
    btn.textContent = t("\u6279\u51C6\uFF08= Apply\uFF09");
    if (!r.ok) {
      toast(esc(t("Apply \u5931\u8D25\uFF08") + String(r.error || r.status) + t("\uFF09\u2014\u2014\u7A3F\u4FDD\u7559\u8BED\u4E49\uFF0C\u672A\u843D\u5E93\u3002")), "error");
      return;
    }
    ignored.add("draft|" + d.kind + "|" + (d.scope || "global") + "|" + d.name);
    const cardEl = document.querySelector('[data-ap-item="' + apKeyOf(d) + '"]');
    if (cardEl) cardEl.remove();
    ctx2.api.snapshot().then((s) => {
      if (s.ok) ctx2.store.init(s.data);
    });
  }
  async function pubGateDecision(ctx2, btn, grant) {
    const anchor = btn.getAttribute("data-ap-anchor") || "";
    const key = btn.getAttribute(grant ? "data-ap-grant" : "data-ap-reject") || "";
    const evId = btn.getAttribute("data-ap-ev") || "";
    if (!anchor || !key) return;
    const confirmed = await confirmModal({
      title: grant ? t("\u6279\u51C6\u7EE7\u7EED") : t("\u9A73\u56DE\u7EC8\u6B62"),
      body: esc(grant ? t("\u786E\u8BA4\u6279\u51C6\u7EE7\u7EED\uFF1F") : t("\u786E\u8BA4\u9A73\u56DE\u7EC8\u6B62\uFF1F")) + ' <b class="mono">' + esc(anchor) + "</b>" + esc(grant ? t("\uFF1A\u53D1\u5E03 ApprovalGranted \u540E run \u6CBF\u58F0\u660E\u76F8\u4F4D\u5E8F\u5217\u7EED\u884C\u3002") : t("\uFF1A\u53D1\u5E03 ApprovalRejected \u540E run \u5C06\u8FDB\u5165\u7EC8\u6001 Failed\uFF08\u4E0D\u53EF\u6062\u590D\uFF09\u3002")),
      confirmText: grant ? t("\u6279\u51C6") : t("\u9A73\u56DE"),
      danger: !grant
    });
    if (!confirmed) return;
    btn.disabled = true;
    btn.textContent = t("\u53D1\u5E03\u4E2D\u2026");
    const runNo = Number(btn.getAttribute("data-ap-run") || 0) || 0;
    const payload = { type: grant ? "ApprovalGranted" : "ApprovalRejected", involved: anchor, ...runNo > 0 ? { run: runNo } : {} };
    if (evId) payload.dedupKey = "console-" + evId;
    else payload.dedupKey = "console-appr-" + key + "-" + Date.now();
    const r = await ctx2.api.eventPub(payload);
    if (!r.ok) {
      btn.disabled = false;
      btn.textContent = grant ? t("\u6279\u51C6\u7EE7\u7EED") : t("\u9A73\u56DE\u7EC8\u6B62");
      toast(esc(t("\u4E8B\u4EF6\u53D1\u5E03\u5931\u8D25\uFF08") + String(r.error || r.status) + t("\uFF09\u2014\u2014run \u4ECD\u6302\u539F\u76F8\u4F4D\uFF0C\u672A\u88C5\u6210\u529F\u3002")), "error");
      return;
    }
    ignored.add(key);
    const card = btn.closest(".ap-item");
    if (card) card.remove();
    ctx2.api.snapshot().then((s) => {
      if (s.ok) ctx2.store.init(s.data);
    });
  }
  async function proposalDecision(ctx2, p, decision, btn) {
    const isApprove = decision === "approve";
    const fm = isApprove ? "console:approve" : "console:reject";
    const confirmed = await confirmModal({
      title: isApprove ? t("\u6279\u51C6\u63D0\u6848") : t("\u9A73\u56DE\u63D0\u6848"),
      body: esc((isApprove ? t("\u786E\u8BA4\u6279\u51C6\u63D0\u6848\uFF08Apply console:approve\uFF09") : t("\u786E\u8BA4\u9A73\u56DE\u63D0\u6848\uFF08Apply console:reject\uFF09")) + " " + String(p.name) + "\uFF1F"),
      confirmText: isApprove ? t("\u6279\u51C6") : t("\u9A73\u56DE"),
      danger: !isApprove
    });
    if (!confirmed) return;
    btn.disabled = true;
    btn.textContent = t("\u5224\u5B9A\u4E2D\u2026");
    const r = await ctx2.api.apply({
      kind: "Proposal",
      scope: p.scope || "global",
      name: p.name,
      fromDraft: true,
      fieldManager: fm,
      reasonRef: "console:" + decision + ":" + String(p.name || "")
    });
    btn.disabled = false;
    btn.textContent = isApprove ? t("\u6279\u51C6\uFF08= Apply\uFF09") : t("\u9A73\u56DE");
    if (!r || !r.ok) {
      toast(esc(t("\u63D0\u6848\u5224\u5B9A\u5931\u8D25\uFF08") + String(r && (r.error || r.status) || t("\u7F51\u7EDC")) + t("\uFF09\u2014\u2014\u63D0\u6848\u72B6\u6001\u672A\u7FFB\u8F6C\uFF0C\u53EF\u91CD\u8BD5\u3002")), "error");
      return;
    }
    ignored.add("proposal|" + (p.scope || "global") + "|" + String(p.name));
    const card = btn.closest(".ap-item");
    if (card) card.remove();
    ctx2.api.snapshot().then((s) => {
      if (s.ok) ctx2.store.init(s.data);
    });
  }
  async function hydrateLedger(ctx2) {
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const slot = document.querySelector("[data-ap-ledger]");
    if (!slot) return;
    const r = await ctx2.api.approvals(20);
    if (!el__isConnected(slot) || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
    if (!r.ok) {
      slot.innerHTML = t('<div class="small faint">\u5BA1\u6279\u8BB0\u5F55\u6682\u4E0D\u53EF\u8BFB\uFF08') + esc(String(r.error || r.status)) + t("\uFF09\u3002</div>");
      return;
    }
    const rows = r.data && Array.isArray(r.data.approvals) ? r.data.approvals : [];
    if (!rows.length) {
      slot.innerHTML = t('<div class="small faint">\u6682\u65E0\u5DF2\u51B3\u8BB0\u5F55\uFF08\u6279\u51C6/\u9A73\u56DE\u540E\u6B64\u5904\u663E\u793A\uFF09\u3002</div>');
      return;
    }
    slot.innerHTML = t('<h2 class="sec-title">\u5DF2\u51B3\u8BB0\u5F55</h2><div style="overflow-x:auto"><table class="tbl"><thead><tr>') + t("<th>\u65F6\u523B</th><th>\u5224\u5B9A</th><th>\u5BF9\u8C61</th><th>\u5224\u5B9A\u4E3B\u4F53</th><th>rv</th></tr></thead><tbody>") + rows.map((x) => '<tr><td class="mono small">' + esc(String(x.ts || "").replace("T", " ").slice(0, 19)) + "</td><td>" + chip(String(x.decision || ""), x.decision === "granted" ? "ok" : "esc") + '</td><td class="mono small">' + (involvedRoute(x.obj) ? '<span class="lk" data-go="' + esc(involvedRoute(x.obj)) + '">' + esc(String(x.obj || "")) + "</span>" : esc(String(x.obj || ""))) + '</td><td class="mono small">' + esc(String(x.judgedBy || "anonymous")) + '</td><td class="mono small">' + esc(x.rv ? "r" + x.rv : "\u2014") + "</td></tr>").join("") + "</tbody></table></div>";
  }
  function el__isConnected(x) {
    return !!(x && x.isConnected);
  }
  function inboxHtml(ctx2) {
    const inbox = ctx2.store.inbox();
    if (!inbox) return t('<div class="card"><div class="card-b"><div class="small faint">\u6536\u4EF6\u7BB1\u6682\u4E0D\u53EF\u7528\uFF08\u6295\u5F71\u7F3A\u5E2D\uFF09\u3002</div></div></div>');
    const items = Array.isArray(inbox.items) ? inbox.items : [];
    if (!items.length && !Number(inbox.total)) {
      return t('<div class="card"><div class="card-b"><div class="small faint">\u6536\u4EF6\u7BB1\u4E3A\u7A7A\uFF08\u96F6\u5F85\u529E\uFF09\u3002</div></div></div>');
    }
    return '<div class="card">' + items.slice(0, 50).map((x) => {
      const linkRaw = String(x && (x.link || x.ref || x.obj) || "");
      const link2 = linkRaw.replace(/^#\//, "");
      const route = link2 ? involvedRoute(link2) || link2 : "";
      const main = String(x && (x.digest || x.title || x.summary || x.kind) || "\u2014");
      const cls = String(x && x.class || x && x.kind || "");
      const at = relAge(String(x && x.at || "")) || "";
      return '<div class="mini-row"><div class="mr-main"><div>' + esc(trunc(main, 80)) + " " + chip(cls, "mute") + '</div><div class="mr-sub small faint">' + esc(at) + "</div></div>" + (route ? '<span class="ws-chip">' + chipGo(t("\u53BB\u5904\u7406"), route, "acc", true) + "</span>" : "") + "</div>";
    }).join("") + "</div>";
  }
  function approvalsScreen(el, ctx2) {
    const focus = String(ctx2.params && ctx2.params.focus || "").trim();
    const rawItems = pendingItems(ctx2);
    const matched = focus ? rawItems.filter((it) => matchesApprovalFocus(it, focus)) : [];
    const rest = focus ? rawItems.filter((it) => !matchesApprovalFocus(it, focus)) : rawItems;
    const items = matched.concat(rest);
    const focusBanner = focus ? '<div class="card" style="margin-bottom:10px"><div class="card-b small" style="display:flex;gap:8px;align-items:center;flex-wrap:wrap">' + chip(t("\u5B9A\u4F4D\uFF1A") + focus, matched.length ? "acc" : "warn") + "<span>" + esc(matched.length ? t("\u5DF2\u7F6E\u9876\u9AD8\u4EAE\u5BF9\u5E94\u5F85\u5BA1\u9879\uFF08") + matched.length + t(" \u9879\uFF09\u3002") : t("\u5F85\u5BA1\u533A\u6682\u65E0\u5339\u914D\u9879\u2014\u2014\u53EF\u80FD\u5DF2\u51B3\u6216\u5C1A\u672A\u843D\u7A3F\uFF0C\u8BF7\u6838\u67E5\u4E0B\u65B9\u5DF2\u51B3\u8BB0\u5F55\u3002")) + "</span>" + chipGo(t("\u6E05\u9664\u5B9A\u4F4D"), "approvals", "mute", true) + "</div></div>" : "";
    el.innerHTML = '<div class="view-inner">' + screenHead(
      t("\u5BA1\u6279"),
      t("\u5F85\u5BA1\u533A=\u8349\u7A3F\u7F13\u5B58\uFF08DocReview \u7C7B\u540C\u6E90\uFF09\u222A ApprovalNeeded \u5F52\u6863\u884C\u3002\u4E8B\u4EF6\u884C\u6279\u51C6/\u9A73\u56DE=\u5BF9 Workflow \u951A\u53D1\u5E03 ApprovalGranted/ApprovalRejected\uFF08run \u7EED\u884C/\u7EC8\u6001 Failed\uFF09\uFF1B\u8349\u7A3F\u6279\u51C6=Apply\uFF1B\u5747=\u4EBA\u70B9\u786E\u8BA4\u3002"),
      chip(t("\u6279\u91CF\u6279\u51C6\u901A\u8DEF\u5F85\u540E\u7AEF\u8BCD\u5F62\u2014\u2014\u9010\u6761\u6279\u51C6\uFF08\u8BDA\u5B9E\u6CE8\u8BB0\uFF09"), "mute")
    ) + focusBanner + (items.length ? items.map((it, idx) => itemHtml(ctx2, it, idx, !!focus && matchesApprovalFocus(it, focus))).join("") : emptyBox(t("\u6682\u65E0\u5F85\u5BA1\u4E8B\u9879"), "", chipGo(t("\u56DE Welcome"), "welcome", "acc", false))) + t('<div data-ap-ledger style="margin-top:20px"><div class="small faint">\u5DF2\u51B3\u8BB0\u5F55\u73B0\u8BFB\u4E2D\u2026</div></div>') + t('<h2 class="sec-title">\u6536\u4EF6\u7BB1\uFF08\u5F85\u4F60\u5904\u7406\uFF09</h2>') + inboxHtml(ctx2) + "</div>";
    items.filter((x) => x.kind === "draft").forEach((x) => hydrateDiff(ctx2, x.draft));
    items.filter((x) => x.kind === "run").forEach((x) => {
      const a = verdictAnchorOf(x.run);
      hydrateVerdict(ctx2, {
        kind: a.kind,
        scope: a.scope,
        name: a.name,
        run: x.run.run,
        apVerdictKey: apVerdictKeyOf(x.run, a)
      });
    });
    hydrateLedger(ctx2);
    if (focus && matched.length) {
      const hit = el.querySelector('[data-ap-focused="1"]');
      if (hit && hit.scrollIntoView) hit.scrollIntoView({ behavior: "smooth", block: "center" });
    }
    el.querySelectorAll("[data-ap-approve]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const d = ctx2.store.drafts().find((x) => apKeyOf(x) === btn.dataset.apApprove);
        if (d) approve(ctx2, d, btn);
      });
    });
    el.querySelectorAll("[data-ap-grant]").forEach((btn) => {
      btn.addEventListener("click", () => pubGateDecision(ctx2, btn, true));
    });
    el.querySelectorAll("[data-ap-reject]").forEach((btn) => {
      btn.addEventListener("click", () => pubGateDecision(ctx2, btn, false));
    });
    const propById = new Map((ctx2.store.proposals() || []).filter((p) => p && p.name).map((p) => ["proposal|" + String(p.scope || "global") + "|" + String(p.name), p]));
    el.querySelectorAll("[data-ap-prop]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const p = propById.get(btn.getAttribute("data-ap-prop") || "");
        if (p) proposalDecision(ctx2, p, String(btn.getAttribute("data-ap-prop-decision") || "approve"), btn);
      });
    });
    el.querySelectorAll("[data-ap-ignore]").forEach((btn) => {
      btn.addEventListener("click", () => {
        ignored.add(btn.dataset.apIgnore);
        const card = btn.closest(".ap-item");
        if (card) card.remove();
      });
    });
    el.querySelectorAll("[data-go]").forEach((x) => {
      x.addEventListener("click", () => ctx2.go(x.dataset.go));
    });
  }
  function render2(el, ctx2) {
    approvalsScreen(el, ctx2);
  }

  // src/screens/welcome.js
  function taskPhaseOf(x) {
    const s = x && x.status;
    return String(s && (s.phase || s) || x && x.phase || "");
  }
  function statCard(label, num, sub, route, cls) {
    return '<div class="stat ' + esc(cls || "") + '" data-go="' + esc(route) + '" role="link"><div class="st-num">' + esc(fmtNum(num)) + '</div><div class="st-label">' + esc(label) + '<span class="small faint">' + esc(sub || "") + "</span></div></div>";
  }
  function loopCardHtml(data) {
    const d = data && typeof data === "object" ? data : {};
    const fields = [
      [t("\u5FAA\u73AF"), d.loop != null ? d.loop : d.id],
      [t("\u76F8\u4F4D"), d.phase || d.state],
      [t("\u8D1F\u8D23"), d.agent || d.route],
      [t("\u5F00\u59CB\u4E8E"), d.startedAt ? relAge(d.startedAt) || String(d.startedAt) : d.started_at ? relAge(d.started_at) || String(d.started_at) : null],
      [t("\u89E6\u53D1"), d.trigger || d.reason]
    ].filter(([, v]) => v != null && v !== "");
    const chips = fields.map(([k, v]) => chip(k + ": " + v, "mute")).join("");
    return '<div class="card report-loop"><div class="card-b"><div class="report-loop-top">' + chip(t("\u5F53\u524D\u5FAA\u73AF"), "acc") + chips + '<span class="report-loop-acts"><span data-ignite-note></span><button class="btn btn-secondary sm" data-ignite="1">' + esc(t("\u89E6\u53D1\u4E8B\u4EF6")) + "</button></span></div></div></div>";
  }
  function bindIgnite(el, ctx2) {
    const btn = el.querySelector("[data-ignite]");
    if (!btn) return;
    btn.addEventListener("click", async () => {
      if (btn.disabled) return;
      const note2 = el.querySelector("[data-ignite-note]");
      btn.disabled = true;
      btn.textContent = t("\u70B9\u706B\u4E2D\u2026");
      const r = await ctx2.api.eventPub({
        type: "LoopRequested",
        involved: "ServiceDesc/core/core_system",
        dedupKey: "console-ignite-" + Date.now()
      });
      if (!el.isConnected) return;
      if (r && r.ok) {
        if (note2) note2.innerHTML = chip(t("\u5DF2\u6CE8\u5165 LoopRequested\uFF08queue \u8BED\u4E49\u2014\u2014\u6865\u63A5\u9762\u62C9\u8D77\u5FAA\u73AF\uFF09"), "ok");
        setTimeout(() => {
          if (el.isConnected) hydrateLoop(el, ctx2);
        }, 3e3);
      } else {
        btn.disabled = false;
        btn.textContent = t("\u89E6\u53D1\u4E8B\u4EF6");
        if (note2) note2.innerHTML = chip(
          t("\u70B9\u706B\u5931\u8D25\uFF08") + String(r && r.status || "\u2014") + (r && r.error ? " " + String(r.error) : "") + t("\uFF09\u2014\u2014\u53EF\u91CD\u8BD5"),
          "danger"
        );
      }
    });
  }
  async function hydrateLoop(el, ctx2) {
    const box = el.querySelector("[data-loop-box]");
    if (!box) return;
    const r = await ctx2.api.loopCurrent();
    if (!el.isConnected || box.getAttribute("data-loop-box") === "stale") return;
    if (!r.ok) {
      box.innerHTML = '<div class="card report-loop" data-report-loop="degraded"><div class="card-b"><div class="report-loop-top">' + chip(t("\u5F53\u524D\u5FAA\u73AF"), "mute") + t('<span class="small faint">\u5FAA\u73AF\u72B6\u6001\u6682\u4E0D\u53EF\u8FBE\uFF08') + esc(String(r.error || r.status)) + t("\uFF09\u2014\u2014\u4E0D\u5192\u5145\u65E0\u5FAA\u73AF\u3002</span></div></div></div>");
      return;
    }
    if (!r.data || !r.data.loop && !r.data.id && !r.data.phase) {
      box.innerHTML = '<div class="card report-loop" data-report-loop="idle"><div class="card-b"><div class="report-loop-top">' + chip(t("\u5F53\u524D\u65E0\u8FDB\u884C\u4E2D\u5FAA\u73AF"), "mute") + t('<span class="small faint">\u672C\u5468\u671F\u6682\u65E0\u5FAA\u73AF\u8BB0\u5F55\u3002\u70B9\u706B\u4E0B\u4E00\u8F6E\uFF1A<code class="mono">corectl loop</code></span></div></div></div>');
      return;
    }
    box.innerHTML = loopCardHtml(r.data);
    bindIgnite(el, ctx2);
  }
  function zoneHead(label, route) {
    return '<div class="zone-h"><span class="zone-t">' + esc(label) + "</span>" + (route ? chipGo(t("\u5168\u90E8"), route, "mute", false) : "") + "</div>";
  }
  function runningWorkflowsHtml(ctx2, scope) {
    const runs = ctx2.store.runs().filter((r) => {
      if (scope) {
        let hay = "";
        try {
          hay = JSON.stringify([r && r.scope, r && r.ownerScope, r && r.involved, r && r.route, r && r.workflow]);
        } catch {
          hay = "";
        }
        if (!hay.includes(scope)) return false;
      }
      const st = String(r && (r.status || r.state) || "");
      return st && st !== "done" && st !== "failed" && st !== "canceled";
    }).slice(0, 8);
    const rows = runs.map((r) => {
      const wf = String(r.workflow || r.wf || "\u2014");
      const phase = String(r.phase || r.state || "\u2014");
      const st = String(r.status || "");
      const started = r.startedAt ? relAge(r.startedAt) || "" : "";
      const route = r.workflow ? "wf/" + encodeURIComponent(wf) + "/r/" + encodeURIComponent(String(r.run != null ? r.run : "")) : "workflows";
      return '<div class="dash-row" data-go="' + esc(route) + '" role="link"><div class="dr-main"><b class="mono">' + esc(trunc(wf, 26)) + '</b><span class="small faint">#' + esc(String(r.run != null ? r.run : "\u2014")) + '</span></div><div class="dr-side">' + chip(phase, st === "running" ? "acc" : "mute") + /* service 节点挂起面（2026-11-27 session 裁决）：suspendWhy=service → service 会话执行中
       * 徽章（非审批挂起——审批一等行不收，workflow.md §1.3）。 */
      (String(r.suspendWhy || "") === "service" ? chip(t("service \u4F1A\u8BDD\u6267\u884C\u4E2D"), "llm") : "") + (started ? '<span class="small faint">' + esc(started) + "</span>" : "") + "</div></div>";
    }).join("");
    return '<div class="card">' + (rows || '<div class="ob-empty">' + esc(t("\u5F53\u524D\u65E0\u8FDB\u884C\u4E2D\u5FAA\u73AF")) + "</div>") + "</div>";
  }
  function runOwnerScope(r) {
    const inv = String(r && (r.involved || r.route || "") || "");
    const m = /^Workflow\/([^/\s]+)\//.exec(inv);
    if (m) return m[1];
    return String(r && (r.scope || r.ownerScope) || "") || "global";
  }
  function pendingRunsOf(ctx2) {
    return ctx2.store.runs().filter((r) => r && (String(r.status || r.state || "") === "suspended" || String(r.phase || "") === "AwaitingApproval")).filter((r) => (
      /* service 节点挂起（suspendWhy=service）非审批态——session 执行中，审批一等行不收
       *（完成信号 ServiceSessionCompleted 续行；2026-11-27 session 裁决）。 */
      String(r.suspendWhy || "") !== "service"
    ));
  }
  function pendingRunRowHtml(x, infoHtml = "") {
    const btnAttrs = (t2) => " data-appr-" + t2 + '="1" data-appr-wf="' + esc(x.wf) + '" data-appr-scope="' + esc(x.scope) + '" data-appr-run="' + esc(x.run) + '" data-appr-key="' + esc(x.key) + '"';
    return '<div class="dash-row appr-row" data-appr-row><div class="dr-main">' + chip(t("\u5F85\u6279"), "warn") + '<b class="mono">' + esc(trunc(x.wf, 22)) + '</b><span class="small faint">#' + esc(x.run) + "</span>" + chip(x.phase, "warn") + (x.started ? '<span class="small faint">' + esc(x.started) + "</span>" : "") + (infoHtml ? '<div class="small" style="flex-basis:100%;margin-top:4px">' + infoHtml + "</div>" : "") + '</div><div class="appr-acts">' + /* V-04（2026-11-26 概念可见性批）：待批行加「运行详情」→ wf/:name/r/:n（此前泛链 absent——
     * 人批准前看不到 run 全貌）。 */
    (x.wf && x.run != null && x.run !== "" ? '<button class="btn btn-ghost sm" data-go="wf/' + encodeURIComponent(String(x.wf)) + "/r/" + encodeURIComponent(String(x.run)) + '">' + esc(t("\u8FD0\u884C\u8BE6\u60C5")) + "</button>" : "") + '<button class="btn btn-chip sm"' + btnAttrs("grant") + ">" + esc(t("\u6279\u51C6\u7EE7\u7EED")) + '</button><button class="btn btn-ghost btn-danger sm"' + btnAttrs("reject") + ">" + esc(t("\u9A73\u56DE\u7EC8\u6B62")) + "</button></div></div>";
  }
  function bindApprovalGates(el, ctx2) {
    const act = async (btn, type2) => {
      if (btn.disabled) return;
      const wf = btn.getAttribute("data-appr-wf") || "";
      const scope = btn.getAttribute("data-appr-scope") || "global";
      const run = btn.getAttribute("data-appr-run") || "";
      const key = btn.getAttribute("data-appr-key") || run;
      const grant = type2 === "ApprovalGranted";
      const ok = await confirmModal({
        title: grant ? t("\u6279\u51C6\u7EE7\u7EED") : t("\u9A73\u56DE\u7EC8\u6B62"),
        body: esc(grant ? t("\u786E\u8BA4\u6279\u51C6\u7EE7\u7EED\uFF1Frun \u5C06\u6CBF Workflow \u58F0\u660E\u76F8\u4F4D\u5E8F\u5217\u7EED\u884C\uFF1A") : t("\u786E\u8BA4\u9A73\u56DE\u7EC8\u6B62\uFF1Frun \u5C06\u8FDB\u5165\u7EC8\u6001 Failed\uFF08\u4E0D\u53EF\u6062\u590D\uFF09\uFF1A")) + ' <b class="mono">' + esc(wf) + " #" + esc(run) + "</b>",
        confirmText: grant ? t("\u6279\u51C6") : t("\u9A73\u56DE"),
        danger: !grant
      });
      if (!ok) return;
      btn.disabled = true;
      const runNo = Number(run) || 0;
      const r = await ctx2.api.eventPub({
        type: type2,
        involved: "Workflow/" + (scope || "global") + "/" + wf,
        dedupKey: "console-" + key,
        ...runNo > 0 ? { run: runNo } : {}
      });
      if (r && r.ok) {
        const row = btn.closest("[data-appr-row]");
        if (row) {
          row.innerHTML = '<div class="dr-main">' + chip(grant ? t("\u5DF2\u6279\u51C6\uFF08ApprovalGranted \u5DF2\u53D1\u5E03\uFF09") : t("\u5DF2\u9A73\u56DE\uFF08ApprovalRejected \u5DF2\u53D1\u5E03\uFF09"), "ok") + '<span class="small faint">' + esc(t("\u5224\u5B9A\u5DF2\u843D\u5BA1\u6279\u53F0\u8D26\u2014\u2014\u5FEB\u7167\u91CD\u57FA\u540E\u884C\u96C6\u5237\u65B0\u3002")) + '</span></div><div class="dr-side">' + chipGo(t("\u67E5\u770B\u5DF2\u51B3\u53F0\u8D26 \u2192"), "approvals", "acc", true, t("\u5BA1\u6279\u9875\u5DF2\u51B3\u53F0\u8D26")) + "</div>";
        }
        ctx2.api.snapshot().then((s) => {
          if (s && s.ok) ctx2.store.init(s.data);
        });
      } else {
        btn.disabled = false;
        const verb = type2 === "ApprovalGranted" ? t("\u6279\u51C6\u5931\u8D25") : t("\u9A73\u56DE\u5931\u8D25");
        toast(esc(verb + t("\uFF08") + String(r && r.status || "\u2014") + (r && r.error ? " " + String(r.error) : "") + t("\uFF09\u2014\u2014run \u4ECD\u5728\u7B49\u5BA1\u6279\uFF0C\u53EF\u91CD\u8BD5\u3002")), "error");
      }
    };
    el.querySelectorAll("[data-appr-grant]").forEach((btn) => btn.addEventListener("click", (ev) => {
      ev.stopPropagation();
      act(btn, "ApprovalGranted");
    }));
    el.querySelectorAll("[data-appr-reject]").forEach((btn) => btn.addEventListener("click", (ev) => {
      ev.stopPropagation();
      act(btn, "ApprovalRejected");
    }));
  }
  function pendingApprovalsHtml(ctx2, scope) {
    const runRows = pendingRunsOf(ctx2).slice(0, 8).map((r) => {
      const wf = String(r.workflow || r.wf || "");
      if (!wf) return "";
      if (scope) {
        const sc = runOwnerScope(r);
        if (sc !== scope) return "";
      }
      const run = String(r.run != null ? r.run : "");
      const key = r.eventId != null && r.eventId !== "" ? String(r.eventId) : run;
      return pendingRunRowHtml({
        wf,
        run,
        key,
        scope: runOwnerScope(r),
        phase: String(r.phase || r.state || "AwaitingApproval"),
        started: r.startedAt ? relAge(r.startedAt) || "" : ""
      }, runApprovalInfoHtml(runApprovalInfoOf(ctx2.store, wf, runOwnerScope(r), run, String(r.startedAt || ""))));
    }).filter(Boolean);
    const drafts = ctx2.store.drafts().filter((d) => d && d.kind && d.name).filter((d) => !scope || String(d.scope || "global") === scope).map((d) => ({
      src: "draft",
      t: d.updatedAt || d.ts || null,
      go: "doc/" + encodeURIComponent(d.kind) + "/" + encodeURIComponent(d.scope || "global") + "/" + encodeURIComponent(d.name),
      title: d.kind + "/" + (d.scope || "global") + "/" + d.name,
      tag: t("\u7A3F")
    }));
    const evs = ctx2.store.events().filter((e) => e && (e.type === "ApprovalNeeded" || e.type === "Escalate")).filter((e) => !scope || String(e.involved || "").includes(scope)).slice(0, 8).map((e) => ({
      src: "ev",
      t: e.ts || null,
      go: "event/" + encodeURIComponent(String(e.id ?? "")),
      title: e.type + " \xB7 " + String(e.involved || ""),
      tag: e.type === "ApprovalNeeded" ? t("\u5F85\u5BA1") : t("\u5347\u7EA7")
    }));
    const rest = drafts.concat(evs).sort((a, b) => String(b.t || "").localeCompare(String(a.t || ""))).slice(0, 8).map((x) => '<div class="dash-row" data-go="' + esc(x.go) + '" role="link"><div class="dr-main">' + chip(x.tag, x.src === "draft" ? "warn" : "esc") + '<b class="mono">' + esc(trunc(x.title, 34)) + "</b></div>" + (x.t ? '<div class="dr-side"><span class="small faint">' + esc(relAge(x.t) || "") + "</span></div>" : "") + "</div>").join("");
    const rows = runRows.join("") + rest;
    if (!rows) return '<div class="card"><div class="ob-empty">' + esc(t("\u5F85\u5BA1\u533A\u4E3A\u7A7A")) + "</div></div>";
    return '<div class="card">' + rows + "</div>";
  }
  var AG_SESS = { list: null, at: 0 };
  var AG_SESS_TTL_MS = 6e4;
  async function sessionsCached(ctx2) {
    if (AG_SESS.list && Date.now() - AG_SESS.at < AG_SESS_TTL_MS) return AG_SESS.list;
    const r = await ctx2.api.sessions();
    if (r && r.ok && Array.isArray(r.data && r.data.sessions)) {
      AG_SESS = { list: r.data.sessions, at: Date.now() };
      return AG_SESS.list;
    }
    return null;
  }
  function sessStateCls(st) {
    return st === "escalated" ? "esc" : st === "running" ? "acc" : st === "suspected-stall" ? "warn" : "mute";
  }
  function sessHitOf(list2, name) {
    const byId = list2.find((s) => String(s && s.sessionId || "").includes(name));
    if (byId) return byId;
    return list2.find((s) => {
      try {
        return JSON.stringify([s && s.involved, s && s.route, s && s.agent]).includes(name);
      } catch {
        return false;
      }
    });
  }
  function sessLiveHtml(hit) {
    if (!hit) return chip(t("\u65E0\u4F1A\u8BDD"), "mute");
    const st = String(hit.state || "");
    const sid = String(hit.sessionId || hit.id || "");
    const sub = [
      String(hit.phase || ""),
      trunc(String(hit.lastAction || ""), 18),
      hit.lastBeat ? relAge(hit.lastBeat) || "" : ""
    ].filter(Boolean).join(" \xB7 ");
    const stChip = st ? sid ? chipGo(st, "session/" + encodeURIComponent(sid), sessStateCls(st), true, t("\u4F1A\u8BDD\u8BE6\u60C5\u2014\u2014\u6D3B\u6027\u8D26\u73B0\u8BFB")) : chip(st, sessStateCls(st)) : "";
    return stChip + (sub ? '<span class="small faint">' + esc(sub) + "</span>" : "");
  }
  async function hydrateAgentLive(el, ctx2) {
    const slots = el.querySelectorAll("[data-agent-live]");
    if (!slots.length) return;
    const list2 = await sessionsCached(ctx2);
    if (!el.isConnected) return;
    if (!list2) {
      slots.forEach((slot) => {
        slot.innerHTML = chip(t("\u4F1A\u8BDD\u6D3B\u6027\u6682\u4E0D\u53EF\u8BFB"), "mute");
      });
      return;
    }
    slots.forEach((slot) => {
      const name = String(slot.getAttribute("data-agent-live") || "");
      if (!name) return;
      slot.innerHTML = sessLiveHtml(sessHitOf(list2, name));
    });
  }
  function agentsHtml(ctx2, filterAgents) {
    const objs = ctx2.store.objects();
    let agents = objs.filter((o) => o.kind === "Agent").map((o) => ({ name: o.name, scope: o.scope || "global", route: objRoute("Agent", o.scope || "global", o.name) }));
    if (filterAgents && filterAgents.length) agents = agents.filter((a) => filterAgents.includes(a.name));
    const sessions = ctx2.store.sessions();
    const rows = agents.map((a) => {
      const sess = sessions.find((s) => String(s && (s.agent || s.route) || "").includes(a.name)) || sessions.find((s) => String(s && s.id || "") === a.name);
      const state = sess ? String(sess.state || sess.status || "") : "";
      const sid = sess ? String(sess.sessionId || sess.id || "") : "";
      const liveChip = state ? sid ? chipGo(state, "session/" + encodeURIComponent(sid), state === "escalated" || state === "suspected-stall" ? "esc" : "acc", true, t("\u4F1A\u8BDD\u8BE6\u60C5")) : chip(state, state === "escalated" || state === "suspected-stall" ? "esc" : "acc") : chip(t("\u65E0\u4F1A\u8BDD"), "mute");
      return '<div class="dash-row" data-go="' + esc("chat/" + encodeURIComponent(a.name)) + '" role="link"><div class="dr-main"><b class="mono">' + esc(trunc(a.name, 30)) + '</b><span class="small faint">' + esc(a.scope) + '</span></div><div class="dr-side" data-agent-live="' + esc(a.name) + '">' + liveChip + "</div></div>";
    }).join("");
    if (!rows) return '<div class="card"><div class="ob-empty">' + esc(t("\u6682\u65E0\u5BF9\u8C61")) + "</div></div>";
    return '<div class="card">' + rows + "</div>";
  }
  function dashLayout(ctx2, opts) {
    const store2 = ctx2.store;
    const scope = opts.scope || "";
    const inScope = (o) => !scope || String(o && (o.scope || o.raw && o.raw.scope) || "") === scope;
    const inbox = store2.inbox();
    const inboxItems = inbox ? (Array.isArray(inbox) ? inbox : inbox && inbox.items || []) || [] : [];
    const inboxScoped = scope ? inboxItems.filter((x) => !x || !x.scope || String(x.scope) === scope) : inboxItems;
    const inboxCard = inbox ? statCard(
      t("\u6536\u4EF6\u7BB1"),
      inboxScoped.length,
      t("\u5F85\u4F60\u5904\u7406"),
      "inbox",
      ""
    ) : "";
    const stats = statCard(t("\u5728\u518C\u5BF9\u8C61"), (store2.objects() || []).filter(inScope).length, t("\u5168\u90E8 kind \u5408\u8BA1"), "obj", "") + statCard(t("\u4E8B\u4EF6"), (store2.events() || []).filter((e) => !scope || String(e.involved || "").includes(scope)).length, t("\u5F52\u6863\u884C"), "events", "") + statCard(t("\u5F85\u5BA1"), (store2.drafts() || []).filter((d) => !scope || String(d.scope || "global") === scope).length + (store2.events() || []).filter((e) => e && e.type === "ApprovalNeeded" && (!scope || String(e.involved || "").includes(scope))).length, t("\u5F85\u5BA1\u660E\u7EC6"), "approvals", "warn") + inboxCard + statCard(t("\u8FDB\u884C\u4E2D\u4EFB\u52A1"), (store2.tasks() || []).filter((x) => ["running", "verifying"].includes(taskPhaseOf(x))).filter(inScope).length, t("\u4EFB\u52A1\u4E2D\u5FC3"), "tasks", "");
    return '<div class="view-inner">' + (opts.head || "") + '<div class="dash-stats' + (inboxCard ? "" : " cols-4") + '">' + stats + '</div><div data-loop-box><div class="card report-loop"><div class="card-b"><div class="report-loop-top">' + chip(t("\u5F53\u524D\u5FAA\u73AF"), "mute") + t('<span class="small faint">\u5FAA\u73AF\u72B6\u6001\u73B0\u8BFB\u4E2D\u2026</span>') + '</div></div></div></div><div class="dash-grid"><div class="dash-zone">' + zoneHead(t("\u6B63\u5728\u8DD1\u7684 Workflows"), "workflows") + runningWorkflowsHtml(ctx2, scope) + '</div><div class="dash-zone">' + zoneHead(t("\u6700\u8FD1\u9700\u8981\u7684\u5BA1\u6279"), "approvals") + pendingApprovalsHtml(ctx2, scope) + '</div><div class="dash-zone">' + zoneHead(t("\u6838\u5FC3 Agent"), "tree") + agentsHtml(ctx2, opts.coreAgents) + "</div></div>" + (opts.foot || "") + "</div>";
  }
  function render3(el, ctx2) {
    el.innerHTML = dashLayout(ctx2, {
      head: '<div class="dash-head"><h2 class="page-title">' + esc(t("\u603B\u89C8")) + '</h2><span class="small faint">' + esc(t("\u7CFB\u7EDF\u89C6\u89D2\u5927\u5C4F\u2014\u2014\u6B63\u5728\u8DD1\u7684\u4EFB\u52A1\u3001\u6700\u8FD1\u9700\u8981\u7684\u5BA1\u6279\u3001\u6838\u5FC3 Agent \u4E00\u773C\u53EF\u8BFB")) + "</span></div>"
    });
    hydrateLoop(el, ctx2);
    bindApprovalGates(el, ctx2);
    hydrateAgentLive(el, ctx2);
  }
  function renderService(el, ctx2) {
    const name = String(ctx2.params.name || "");
    const sd = ctx2.store.objects().find((o) => o.kind === "ServiceDesc" && o.name === name);
    el.innerHTML = dashLayout(ctx2, {
      scope: name,
      head: '<div class="dash-head"><h2 class="page-title">' + esc(t("Service")) + " \xB7 " + esc(name) + '</h2><span class="small faint">' + esc(t("Service \u89C6\u89D2\u5927\u5C4F\uFF08scope=" + name + " \u6536\u7A84\u6295\u5F71\u2014\u2014\u4E0E\u7CFB\u7EDF\u5927\u5C4F\u540C\u6784\u5E03\u5C40\uFF09")) + "</span></div>",
      coreAgents: null,
      foot: '<div class="card"><div class="card-b"><div class="small faint" data-sd-note>' + esc(t("ServiceDesc \u6B63\u6587\u73B0\u8BFB\u4E2D\u2026\uFF08\u58F0\u660E agents=agent://<name> \u8BCD\u5F62\u89E3\u6790\uFF09")) + "</div></div></div>"
    });
    hydrateLoop(el, ctx2);
    bindApprovalGates(el, ctx2);
    hydrateAgentLive(el, ctx2);
    (async () => {
      const note2 = el.querySelector("[data-sd-note]");
      if (!note2) return;
      let content = "";
      const cached = sd && ctx2.store.cachedDoc ? ctx2.store.cachedDoc("ServiceDesc", sd && sd.scope || "global", name, sd && sd.rv) : null;
      if (cached && typeof cached.content === "string") {
        content = cached.content;
      } else if (ctx2.api && typeof ctx2.api.object === "function") {
        const r = await ctx2.api.object("ServiceDesc", sd && sd.scope || "global", name);
        if (!el.isConnected) return;
        if (r && r.ok) {
          const data = r.data && typeof r.data === "object" ? r.data : {};
          content = data.content != null ? String(data.content) : "";
          if (sd && ctx2.store.cacheDoc) ctx2.store.cacheDoc("ServiceDesc", sd && sd.scope || "global", name, sd && sd.rv, content);
        }
      }
      const agents = [...content.matchAll(/agent:\/\/([A-Za-z0-9_.-]+)/g)].map((m) => m[1]);
      note2.textContent = sd ? content ? t("ServiceDesc \u5728\u518C\u2014\u2014agents=") + (agents.join(", ") || "\u2014") : t("ServiceDesc \u5728\u518C\u2014\u2014\u6B63\u6587\u6682\u4E0D\u53EF\u8BFB\uFF08agents \u89E3\u6790\u7F3A\u5E2D\uFF0C\u4E0D\u5192\u5145\uFF09") : t("ServiceDesc \u5BF9\u8C61\u4E0D\u5728\u518C\uFF08\u8BCD\u5F62\u6295\u5F71\u81EA\u5FEB\u7167\u2014\u2014\u4E0D\u5192\u5145\u5728\u518C\uFF09");
    })().catch(() => {
    });
  }

  // src/screens/tasks.js
  var PHASE_VIEW = {
    running: { label: t("\u6B63\u5728\u505A"), chip: "acc" },
    verifying: { label: t("\u9A8C\u8BC1\u4E2D"), chip: "acc" },
    queued: { label: t("\u6392\u961F\u4E2D"), chip: "mute" },
    ready: { label: t("\u5F85\u542F\u52A8"), chip: "mute" },
    pending: { label: t("\u7B49\u5F85\u62C9\u8D77"), chip: "mute" },
    blocked: { label: t("\u53D7\u963B"), chip: "warn" },
    failed: { label: t("\u5931\u8D25"), chip: "esc" },
    done: { label: t("\u5DF2\u5B8C\u6210"), chip: "ok" },
    canceled: { label: t("\u5DF2\u53D6\u6D88"), chip: "mute" }
  };
  function statusOf(t2) {
    const s = t2 && t2.status || {};
    const flat = (k) => s[k] != null ? s[k] : t2 ? t2[k] : void 0;
    return {
      phase: flat("phase"),
      attempts: flat("attempts"),
      agentSession: flat("agentSession"),
      sessionId: flat("sessionId"),
      agent: flat("agent"),
      lastAction: flat("lastAction"),
      lastError: flat("lastError"),
      dependsOn: s.dependsOn != null ? s.dependsOn : t2 && t2.dependsOn
    };
  }
  function specOf(t2) {
    return t2 && t2.spec || {};
  }
  function phaseChip(t2) {
    const phase = String(statusOf(t2).phase || "");
    const v = PHASE_VIEW[phase];
    if (!v) return chip(phase || t("\u72B6\u6001\u672A\u77E5"), "mute");
    return chip(v.label, v.chip);
  }
  function humanWait(t2) {
    const s = statusOf(t2);
    return /审批|批准|待审|等待人|approval|approve/i.test(String(s.lastError || "") + " " + String(s.lastAction || ""));
  }
  function groupOf(t2) {
    const phase = String(statusOf(t2).phase || "");
    if (phase === "running" || phase === "verifying") return "doing";
    if (phase === "done" || phase === "canceled") return "done";
    if (phase === "blocked" || phase === "failed") return humanWait(t2) ? "waiting" : "stuck";
    return "queued";
  }
  var GROUPS = [
    { id: "doing", label: t("\u6B63\u5728\u505A") },
    { id: "waiting", label: t("\u7B49\u4EBA") },
    { id: "stuck", label: t("\u5361\u4F4F") },
    { id: "done", label: t("\u5B8C\u6210") },
    { id: "queued", label: t("\u6392\u961F") }
  ];
  function groupTasks(tasks) {
    const out = { doing: [], waiting: [], stuck: [], done: [], queued: [] };
    (tasks || []).forEach((t2) => {
      out[groupOf(t2)].push(t2);
    });
    return out;
  }
  function taskIdOf(t2) {
    if (!t2) return "";
    if (t2.taskId) return String(t2.taskId);
    return (t2.scope || "global") + "/" + String(t2.name || "");
  }
  function taskTitle(t2) {
    const spec = specOf(t2);
    return String(spec.title || t2 && t2.title || t2 && t2.name || t("\u672A\u547D\u540D\u4EFB\u52A1"));
  }
  var cancelReq = /* @__PURE__ */ new Set();
  function cancellable(t2) {
    const phase = String(statusOf(t2).phase || "");
    return !phase || phase === "queued" || phase === "pending";
  }
  function cancelBtnHtml(t2) {
    const id = taskIdOf(t2);
    if (cancelReq.has(id)) return chip(t("\u53D6\u6D88\u8BF7\u6C42\u5DF2\u53D1\u5E03"), "warn");
    if (!cancellable(t2)) return "";
    return '<button class="btn btn-ghost btn-danger sm" data-task-cancel="' + esc(id) + '">' + esc(t("\u53D6\u6D88")) + "</button>";
  }
  async function requestCancel(el, ctx2, t2) {
    const name = String(t2 && t2.name || "");
    const scope = String(t2 && t2.scope || "global") || "global";
    if (!name) {
      toast(esc(t("\u4EFB\u52A1\u7F3A name\u2014\u2014\u65E0\u6CD5\u6784\u9020 Task/<scope>/<name> \u8BCD\u5F62\uFF0C\u53D6\u6D88\u672A\u53D1\u5E03")), "error");
      return;
    }
    const ok = await confirmModal({
      title: t("\u53D6\u6D88\u4EFB\u52A1"),
      body: esc(t("\u53D6\u6D88\u4EFB\u52A1 ") + name + t("\uFF1F\u53D1\u5E03 TaskCancelRequested\uFF08\u8C03\u5EA6\u5668\u5F52\u6863\u9762\u6D88\u8D39\u2014\u2014\u4E0D\u4FDD\u8BC1\u5DF2\u5728\u8DD1\u7684\u8FDB\u7A0B\u7ACB\u523B\u7EC8\u6B62\uFF09")),
      confirmText: t("\u53D6\u6D88\u4EFB\u52A1"),
      danger: true
    });
    if (!ok) return;
    try {
      const r = await ctx2.api.eventPub({
        type: "TaskCancelRequested",
        involved: "Task/" + scope + "/" + name,
        dedupKey: "console-cancel-" + name + "-" + Date.now()
      });
      if (r && r.ok) {
        cancelReq.add(taskIdOf(t2));
        ctx2.api.snapshot().then((s) => {
          if (s && s.ok) ctx2.store.init(s.data);
        }).catch(() => {
        });
        if (el.isConnected) render4(el, ctx2);
      } else {
        toast(esc(t("\u53D6\u6D88\u8BF7\u6C42\u53D1\u5E03\u5931\u8D25\uFF1A") + String(r && (r.error || r.status) || t("\u7F51\u7EDC\u5F02\u5E38"))), "error");
      }
    } catch (err) {
      toast(esc(t("\u53D6\u6D88\u8BF7\u6C42\u53D1\u5E03\u5931\u8D25\uFF1A") + String(err && err.message || err || t("\u7F51\u7EDC\u5F02\u5E38"))), "error");
    }
  }
  function depRoute(d, defaultKind = "") {
    const s = String(d || "").trim();
    if (!s) return "";
    const m = /^(Task|task|Workflow|workflow)\/([^/\s]+)\/([^/\s]+)$/.exec(s);
    if (m) {
      return m[1].toLowerCase() === "task" ? "task/" + encodeURIComponent(m[2] + "/" + m[3]) : "wf/" + encodeURIComponent(m[3]);
    }
    const m2 = /^([A-Za-z0-9_.-]+)\/([A-Za-z0-9_.-]+)$/.exec(s);
    if (m2) {
      if (defaultKind === "workflow") return "wf/" + encodeURIComponent(m2[2]);
      return "task/" + encodeURIComponent(m2[1] + "/" + m2[2]);
    }
    if (defaultKind === "workflow" && /^[A-Za-z0-9_.-]+$/.test(s)) return "wf/" + encodeURIComponent(s);
    return "";
  }
  function depChips(deps, defaultKind = "") {
    return (deps || []).map((d) => {
      const r = depRoute(d, defaultKind);
      return r ? chipGo(String(d), r, "mute", true, t("\u4F9D\u8D56\u2014\u2014\u70B9\u51FB\u8FDB\u8BE6\u60C5")) : chip(String(d), "mute");
    }).join(" ");
  }
  function taskRow(t2) {
    const id = taskIdOf(t2);
    const s = statusOf(t2);
    const spec = specOf(t2);
    const depList = spec.dependsOn || s.dependsOn || [];
    const agent = String(s.agent || spec.agent || "").trim();
    const sess = String(s.agentSession || s.sessionId || "").trim();
    const last = String(s.lastAction || "").trim();
    const unknown = !String(s.phase || "");
    return '<div class="task-row"><div class="task-main"><div class="task-title"><b>' + esc(taskTitle(t2)) + "</b>" + phaseChip(t2) + chip(id, "mute") + (depList.length ? depChips(depList.slice(0, 3), "task") + (depList.length > 3 ? chip("+" + (depList.length - 3), "mute") : "") : "") + (agent ? chip(t("\u8D1F\u8D23 ") + agent, "mute") : "") + '</div><div class="task-sub small">' + esc(trunc(last, 100) || t("\u6700\u8FD1\u52A8\u4F5C\uFF1A\u2014")) + (String(s.lastError || "").trim() ? " " + chip(trunc(String(s.lastError), 60), "esc") : "") + "</div>" + (unknown ? '<div class="task-sub small faint">' + esc(t("\u72B6\u6001\u672A\u77E5\uFF08\u53EF\u53D6\u6D88\u6216\u8FDB\u8BE6\u60C5\u6838\u67E5\uFF09")) + "</div>" : "") + '</div><div class="task-acts">' + (sess ? chipGo(t("\u5BF9\u8BDD"), "session/" + encodeURIComponent(sess), "mute", true, t("\u5173\u8054\u4F1A\u8BDD")) : "") + cancelBtnHtml(t2) + "</div></div>";
  }
  function tasksScreen(el, ctx2) {
    const tasks = ctx2.store.tasks();
    const g = groupTasks(tasks);
    const body = GROUPS.map((grp) => {
      const rows = g[grp.id] || [];
      return '<div class="card"><div class="card-h"><span class="h-title">' + esc(grp.label) + "</span>" + chip(String(rows.length) + t(" \u9879"), rows.length ? "acc" : "mute") + "</div>" + (rows.length ? '<div class="card-b">' + rows.map(taskRow).join("") + "</div>" : '<div class="card-b small faint">' + esc(grp.label) + t("\uFF1A\u6682\u65E0\u3002</div>")) + "</div>";
    }).join("");
    el.innerHTML = '<div class="view-inner">' + screenHead(t("\u4EFB\u52A1\u4E2D\u5FC3"), t("\u5206\u7EC4=\u786E\u5B9A\u6027\u6295\u5F71\uFF08\u72B6\u6001\u5355\u4E00\u5199\u8005=\u8C03\u5EA6\u5668\uFF09\uFF1B\u70B9\u51FB\u884C\u8FDB\u8BE6\u60C5"), chipGo(t("\u5173\u6CE8\u62A5\u544A"), "report", "acc", false)) + '<div class="grid cols-2">' + body + "</div></div>";
    el.querySelectorAll(".task-row").forEach((row, i) => {
      let n = 0;
      for (const grp of GROUPS) {
        const rows = g[grp.id] || [];
        if (i < n + rows.length) {
          const t2 = rows[i - n];
          row.style.cursor = "pointer";
          row.addEventListener("click", (ev) => {
            if (ev.target && ev.target.closest("[data-go],[data-task-cancel]")) return;
            ctx2.go("task/" + encodeURIComponent(taskIdOf(t2)));
          });
          return;
        }
        n += rows.length;
      }
    });
    const byId = new Map(tasks.map((x) => [taskIdOf(x), x]));
    el.querySelectorAll("[data-task-cancel]").forEach((btn) => {
      btn.addEventListener("click", (ev) => {
        ev.stopPropagation();
        const t2 = byId.get(btn.dataset.taskCancel);
        if (t2) requestCancel(el, ctx2, t2);
      });
    });
  }
  function taskPhaseOf2(e) {
    const dk = String(e && e.dedupKey || "");
    if (!/^task:/.test(dk)) return "";
    return dk.slice(dk.lastIndexOf(":") + 1);
  }
  function taskEvRow(e) {
    const phase = taskPhaseOf2(e);
    const pv = PHASE_VIEW[phase];
    const ts = e && (e.ts || e.at || e.time);
    return '<div class="mini-row"><span class="w-ev-dot" data-level="' + esc(String(e && e.level || "info")) + '"></span><div class="mr-main"><div>' + esc(String(e && e.type || t("\u672A\u547D\u540D\u4E8B\u4EF6"))) + (pv ? " " + chip(pv.label, pv.chip) : phase ? " " + chip(phase, "mute") : "") + '</div><div class="mr-sub small mono">' + esc(trunc(String(e && e.dedupKey || e && e.eventId || ""), 64)) + '</div></div><span class="w-ev-time mono" title="' + esc(String(ts || "")) + '">' + esc(relAge(ts) || "\u2014") + "</span></div>";
  }
  async function hydrateTaskTimeline(box, ctx2, id) {
    const r = await ctx2.api.events(60, { involved: "Task/" + id });
    if (!box.isConnected) return;
    if (!r || !r.ok) {
      box.innerHTML = '<div class="small faint">' + esc(t("\u65F6\u95F4\u7EBF\u62C9\u53D6\u5931\u8D25\uFF1A")) + esc(String(r && (r.error || r.status) || t("\u7F51\u7EDC\u5F02\u5E38"))) + "</div>";
      return;
    }
    const d = r && r.data || {};
    const rows = Array.isArray(d) ? d : Array.isArray(d.events) ? d.events : [];
    box.innerHTML = rows.length ? rows.map(taskEvRow).join("") : '<div class="small faint">' + esc(t("\u5F52\u6863\u7A97\u5185\u6682\u65E0\u672C\u4EFB\u52A1\u4E8B\u4EF6\u3002")) + "</div>";
  }
  function taskDetailScreen(el, ctx2, id) {
    const tasks = ctx2.store.tasks();
    const t2 = tasks.find((x) => taskIdOf(x) === String(id));
    const head = screenHead(t("\u4EFB\u52A1\u8BE6\u60C5"), "id=" + String(id), chipGo(t("\u2190 \u4EFB\u52A1\u4E2D\u5FC3"), "tasks", "mute", false));
    if (!t2) {
      el.innerHTML = '<div class="view-inner">' + head + emptyBox(t("\u4EFB\u52A1\u4E0D\u5728\u518C"), t("\u6682\u65E0\u6B64\u4EFB\u52A1\uFF08\u53EF\u80FD\u5DF2\u5B8C\u6210\u5F52\u6863\u6216 id \u6709\u8BEF\uFF09\u3002")) + "</div>";
      return;
    }
    const s = statusOf(t2);
    const spec = specOf(t2);
    const rows = [
      [t("\u76F8\u4F4D"), phaseChip(t2)],
      [t("\u6807\u9898"), esc(taskTitle(t2))],
      ["id", esc(taskIdOf(t2))]
    ];
    if (spec.agent || s.agent) {
      const ag = String(s.agent || spec.agent).trim();
      const seg = ag.split("/");
      const scope = seg.length > 1 ? seg[0] : "global";
      const name = seg.length > 1 ? seg.slice(1).join("/") : ag;
      rows.push([t("\u8D1F\u8D23"), chipGo(ag, objRoute("Agent", scope, name), "mute", true, t("\u6253\u5F00 Agent \u5BF9\u8C61\u9875"))]);
    }
    const sessId = String(s.agentSession || s.sessionId || "").trim();
    if (sessId) {
      const sessTarget = sessionTargetOf(sessId);
      rows.push([t("\u4F1A\u8BDD"), chipGo(trunc(sessId, 24), sessTarget, "acc", true, t("\u5173\u8054\u4F1A\u8BDD"))]);
    }
    if (!String(s.phase || "")) rows.push([t("\u6CE8\u8BB0"), esc(t("\u72B6\u6001\u672A\u77E5\uFF08\u53EF\u53D6\u6D88\u6216\u6838\u67E5\u8C03\u5EA6\u65E5\u5FD7\uFF09"))]);
    const deps = spec.dependsOn || s.dependsOn || [];
    if (deps.length) rows.push([t("\u4F9D\u8D56"), depChips(deps, "task")]);
    if (cancelReq.has(taskIdOf(t2))) rows.push([t("\u53D6\u6D88"), chip(t("\u53D6\u6D88\u8BF7\u6C42\u5DF2\u53D1\u5E03"), "warn")]);
    const lastA = String(s.lastAction || "").trim();
    const lastE = String(s.lastError || "").trim();
    const sessLive = sessId ? (ctx2.store.sessions() || []).find((x) => String(x && x.sessionId || "") === sessId) : null;
    const sessLiveHtml2 = sessId ? '<div class="mini-row"><span class="w-ev-dot" data-level="' + esc(/escalated|suspected-stall/.test(String(sessLive && sessLive.state)) ? "error" : "info") + '"></span><div class="mr-main"><div>' + esc(t("\u5173\u8054\u4F1A\u8BDD\u73B0\u52BF")) + " " + (sessLive ? chip(String(sessLive.state || t("\u5728\u518C")), /escalated|suspected-stall/.test(String(sessLive.state)) ? "esc" : "acc") : chip(t("\u4E0D\u5728\u5FC3\u8DF3\u5728\u518C"), "mute")) + (sessLive && sessLive.phase ? chip(String(sessLive.phase), "mute") : "") + (sessLive && sessLive.lastAction ? '<span class="small"> ' + esc(trunc(String(sessLive.lastAction), 80)) + "</span>" : "") + '</div><div class="mr-sub small mono">' + esc(trunc(sessId, 48)) + (sessLive && sessLive.lastBeat ? " \xB7 " + esc(t("\u6700\u8FD1\u5FC3\u8DF3 ")) + esc(relAge(sessLive.lastBeat)) : "") + '</div></div><span class="w-ev-time">' + chipGo(t("\u5BF9\u8BDD"), "session/" + encodeURIComponent(sessId), "mute", true, t("\u5173\u8054\u4F1A\u8BDD")) + "</span></div>" : "";
    const resultHtml = lastA || lastE ? '<div style="margin:8px 0">' + (lastA ? '<div class="small" style="white-space:pre-wrap;word-break:break-word"><b>' + esc(t("\u7ED3\u679C\u6CE8\u8BB0")) + '</b> <span title="' + esc(lastA) + '">' + esc(trunc(lastA, 4e3)) + "</span></div>" : "") + (lastE ? '<div style="margin-top:6px">' + chip(trunc(lastE, 160), "esc") + "</div>" : "") + "</div>" : "";
    el.innerHTML = '<div class="view-inner">' + head + '<div class="card"><div class="card-b">' + kv(rows) + (cancellable(t2) && !cancelReq.has(taskIdOf(t2)) ? '<div style="margin-top:10px"><button class="btn btn-ghost btn-danger sm" data-task-cancel="' + esc(taskIdOf(t2)) + '">' + esc(t("\u53D6\u6D88")) + '</button><span class="small faint" style="margin-left:8px">' + esc(t("\u53D1\u5E03 TaskCancelRequested\u2014\u2014\u8C03\u5EA6\u5668\u5F52\u6863\u9762\u6D88\u8D39")) + "</span></div>" : "") + '</div></div><div class="card" style="margin-top:12px"><div class="card-h"><span class="h-title">' + esc(t("\u6700\u8FD1\u52A8\u4F5C")) + "</span>" + chip(t("\u7ED3\u679C \xB7 \u4F1A\u8BDD\u73B0\u52BF \xB7 \u751F\u547D\u5468\u671F\u65F6\u95F4\u7EBF"), "mute") + '</div><div class="card-b">' + resultHtml + sessLiveHtml2 + '<div class="small faint" style="margin-top:6px">' + esc(t("\u751F\u547D\u5468\u671F\uFF08\u4E8B\u4EF6\u5F52\u6863\u73B0\u8BFB\uFF09")) + '</div><div data-task-tl="' + esc(taskIdOf(t2)) + '"><div class="small faint">' + esc(t("\u73B0\u8BFB\u4E2D\u2026")) + "</div></div></div></div></div>";
    el.querySelectorAll("[data-task-cancel]").forEach((btn) => {
      btn.addEventListener("click", () => requestCancel(el, ctx2, t2));
    });
    const tlBox = el.querySelector("[data-task-tl]");
    if (tlBox) hydrateTaskTimeline(tlBox, ctx2, taskIdOf(t2));
  }
  var SEV_VIEW = { critical: { label: t("\u4E25\u91CD"), chip: "esc" }, high: { label: t("\u91CD\u8981"), chip: "warn" }, medium: { label: t("\u4E00\u822C"), chip: "mute" } };
  function sevView(sev) {
    return SEV_VIEW[String(sev || "")] || { label: String(sev || t("\u672A\u5206\u7EA7")), chip: "mute" };
  }
  function sevRank(sev) {
    const s = String(sev || "");
    return s === "critical" ? 0 : s === "high" ? 1 : s === "medium" ? 2 : 3;
  }
  function sortAttention(list2) {
    return (list2 || []).slice().sort((a, b) => sevRank(a && a.severity) - sevRank(b && b.severity));
  }
  function evidenceRoute(ref) {
    const r = String(ref || "").trim();
    if (!r) return "";
    if (r.startsWith("#/")) {
      const segs = r.slice(2).split("/").filter(Boolean);
      if (!segs.length) return "";
      if (segs[0] === "inbox") return "inbox";
      if (segs[0] === "approvals") return "approvals";
      if (segs[0] === "tasks") return "tasks";
      if (["s", "wf"].includes(segs[0]) && segs.length >= 2) return segs.map((x) => encodeURIComponent(x)).join("/");
      if (["event", "task", "session"].includes(segs[0]) && segs.length >= 2) return segs[0] + "/" + encodeURIComponent(segs.slice(1).join("/"));
      if (segs[0] === "obj" && segs.length >= 4) return segs.map((x) => encodeURIComponent(x)).join("/");
      return "";
    }
    return involvedRoute(r);
  }
  function attentionCard(a) {
    if (!a) return "";
    const sev = sevView(a.severity);
    const evs = (a.evidence || []).map((ev) => {
      const ref = String(ev && ev.ref || "");
      const route2 = evidenceRoute(ref);
      return route2 ? chipGo(ref, route2, "mute", true, t("\u8BC1\u636E\u2014\u2014\u70B9\u51FB\u8FDB\u5165")) : chip(ref || ev && ev.kind || t("\u8BC1\u636E"), "mute");
    }).filter(Boolean);
    const action = a.action || {};
    const route = action.target ? String(action.target).replace(/^#\/?/, "") : "";
    return '<div class="att-card sev-' + esc(String(a.severity || "unknown")) + '"><div class="att-h">' + chip(sev.label, sev.chip) + (a.category ? chip(a.category, "mute") : "") + "<b>" + esc(String(a.title || t("\u672A\u547D\u540D\u5173\u6CE8\u9879"))) + "</b></div>" + (a.why ? '<div class="att-why small">' + esc(String(a.why)) + "</div>" : "") + '<div class="att-ev">' + (evs.length ? evs.join(" ") : t('<span class="small faint">\u8BC1\u636E\uFF1A\u2014</span>')) + "</div>" + (route ? '<div class="att-acts"><button class="btn btn-secondary sm" data-go="' + esc(route) + '">' + esc(String(action.label || t("\u53BB\u5904\u7406"))) + "</button></div>" : "") + "</div>";
  }
  function reportFooter(body) {
    const b = body || {};
    const dq = b.dataQuality || {};
    const missing = dq.missing;
    const missText = Array.isArray(missing) ? missing.map((m) => String(m)).join("\u3001") : String(missing || "");
    const gb = String(b.generatedBy || "");
    const gbLabel = gb === "llm" ? t("\u667A\u80FD\u5F52\u7EB3") : gb === "rules" ? t("\u57FA\u7840\u6C47\u603B") : gb || t("\u751F\u6210\u65B9\u5F0F\u672A\u6CE8\u660E");
    return t('<div class="report-foot small faint">\u62A5\u544A\u8BF4\u660E\uFF1A') + esc(gbLabel) + t(" \xB7 \u68C0\u67E5\u8303\u56F4 ") + esc(String(b.scope || "global")) + (b.asOf ? t(" \xB7 \u622A\u81F3 ") + esc(String(b.asOf)) : "") + (missText ? t(" \xB7 \u7F3A\u5E2D\u6570\u636E\uFF1A") + esc(missText) : "") + t(" \xB7 \u62A5\u544A\u4E3A\u6D3E\u751F\u89C6\u56FE\u53EF\u968F\u65F6\u91CD\u5EFA\uFF0C\u4E0D\u843D\u5BF9\u8C61\u8D26\u3002</div>");
  }
  function reportScreen(el, ctx2) {
    const st = ctx2.store;
    let att = st.attention();
    let degradedNote = "";
    if (!att.length) {
      const evs = st.events().filter((e) => e && (e.type === "ApprovalNeeded" || e.type === "Escalate"));
      att = evs.map((e) => ({
        severity: e.type === "Escalate" ? "high" : "medium",
        category: e.type,
        title: String(e.summary || e.type),
        evidence: e.involved ? [{ ref: String(e.involved) }] : []
      }));
      if (!evs.length) degradedNote = t('<div class="small faint" style="margin-bottom:10px">\u5F53\u524D\u6CA1\u6709\u9700\u8981\u4F60\u5173\u6CE8\u7684\u4E8B\u9879\u3002</div>');
      else degradedNote = t('<div class="small faint" style="margin-bottom:10px">\u4EE5\u4E0B\u4E8B\u9879\u6765\u81EA\u6700\u8FD1\u4E8B\u4EF6\u3002</div>');
    }
    const sorted = sortAttention(att);
    el.innerHTML = '<div class="view-inner">' + screenHead(t("\u5173\u6CE8\u62A5\u544A"), t("severity \u6392\u5E8F \xB7 \u8BC1\u636E\u53EF\u70B9 \xB7 \u6570\u636E\u53EF\u4FE1\u5EA6\u663E\u5F0F"), chipGo(t("\u4EFB\u52A1\u4E2D\u5FC3"), "tasks", "mute", false)) + degradedNote + '<div data-loop-box="pending"></div>' + (sorted.length ? '<div class="att-list">' + sorted.map(attentionCard).join("") + "</div>" : emptyBox(t("\u672C\u7A97\u53E3\u65E0\u5173\u6CE8\u9879"), t("\u5F53\u524D\u6CA1\u6709\u9700\u8981\u5173\u6CE8\u7684\u98CE\u9669\u9879\u3002"))) + reportFooter(st.get("attention") && st.get("attention").dataQuality) + "</div>";
    hydrateLoop(el, ctx2);
  }
  function render4(el, ctx2) {
    if (ctx2.params && ctx2.params.id != null && ctx2.params.id !== "") taskDetailScreen(el, ctx2, String(ctx2.params.id));
    else tasksScreen(el, ctx2);
  }
  function renderReport(el, ctx2) {
    reportScreen(el, ctx2);
  }

  // ../../../web/node_modules/marked/lib/marked.esm.js
  function _getDefaults() {
    return {
      async: false,
      breaks: false,
      extensions: null,
      gfm: true,
      hooks: null,
      pedantic: false,
      renderer: null,
      silent: false,
      tokenizer: null,
      walkTokens: null
    };
  }
  var _defaults = _getDefaults();
  function changeDefaults(newDefaults) {
    _defaults = newDefaults;
  }
  var noopTest = { exec: () => null };
  function edit(regex, opt = "") {
    let source = typeof regex === "string" ? regex : regex.source;
    const obj = {
      replace: (name, val) => {
        let valSource = typeof val === "string" ? val : val.source;
        valSource = valSource.replace(other.caret, "$1");
        source = source.replace(name, valSource);
        return obj;
      },
      getRegex: () => {
        return new RegExp(source, opt);
      }
    };
    return obj;
  }
  var other = {
    codeRemoveIndent: /^(?: {1,4}| {0,3}\t)/gm,
    outputLinkReplace: /\\([\[\]])/g,
    indentCodeCompensation: /^(\s+)(?:```)/,
    beginningSpace: /^\s+/,
    endingHash: /#$/,
    startingSpaceChar: /^ /,
    endingSpaceChar: / $/,
    nonSpaceChar: /[^ ]/,
    newLineCharGlobal: /\n/g,
    tabCharGlobal: /\t/g,
    multipleSpaceGlobal: /\s+/g,
    blankLine: /^[ \t]*$/,
    doubleBlankLine: /\n[ \t]*\n[ \t]*$/,
    blockquoteStart: /^ {0,3}>/,
    blockquoteSetextReplace: /\n {0,3}((?:=+|-+) *)(?=\n|$)/g,
    blockquoteSetextReplace2: /^ {0,3}>[ \t]?/gm,
    listReplaceTabs: /^\t+/,
    listReplaceNesting: /^ {1,4}(?=( {4})*[^ ])/g,
    listIsTask: /^\[[ xX]\] /,
    listReplaceTask: /^\[[ xX]\] +/,
    anyLine: /\n.*\n/,
    hrefBrackets: /^<(.*)>$/,
    tableDelimiter: /[:|]/,
    tableAlignChars: /^\||\| *$/g,
    tableRowBlankLine: /\n[ \t]*$/,
    tableAlignRight: /^ *-+: *$/,
    tableAlignCenter: /^ *:-+: *$/,
    tableAlignLeft: /^ *:-+ *$/,
    startATag: /^<a /i,
    endATag: /^<\/a>/i,
    startPreScriptTag: /^<(pre|code|kbd|script)(\s|>)/i,
    endPreScriptTag: /^<\/(pre|code|kbd|script)(\s|>)/i,
    startAngleBracket: /^</,
    endAngleBracket: />$/,
    pedanticHrefTitle: /^([^'"]*[^\s])\s+(['"])(.*)\2/,
    unicodeAlphaNumeric: /[\p{L}\p{N}]/u,
    escapeTest: /[&<>"']/,
    escapeReplace: /[&<>"']/g,
    escapeTestNoEncode: /[<>"']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/,
    escapeReplaceNoEncode: /[<>"']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/g,
    unescapeTest: /&(#(?:\d+)|(?:#x[0-9A-Fa-f]+)|(?:\w+));?/ig,
    caret: /(^|[^\[])\^/g,
    percentDecode: /%25/g,
    findPipe: /\|/g,
    splitPipe: / \|/,
    slashPipe: /\\\|/g,
    carriageReturn: /\r\n|\r/g,
    spaceLine: /^ +$/gm,
    notSpaceStart: /^\S*/,
    endingNewline: /\n$/,
    listItemRegex: (bull) => new RegExp(`^( {0,3}${bull})((?:[	 ][^\\n]*)?(?:\\n|$))`),
    nextBulletRegex: (indent) => new RegExp(`^ {0,${Math.min(3, indent - 1)}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`),
    hrRegex: (indent) => new RegExp(`^ {0,${Math.min(3, indent - 1)}}((?:- *){3,}|(?:_ *){3,}|(?:\\* *){3,})(?:\\n+|$)`),
    fencesBeginRegex: (indent) => new RegExp(`^ {0,${Math.min(3, indent - 1)}}(?:\`\`\`|~~~)`),
    headingBeginRegex: (indent) => new RegExp(`^ {0,${Math.min(3, indent - 1)}}#`),
    htmlBeginRegex: (indent) => new RegExp(`^ {0,${Math.min(3, indent - 1)}}<(?:[a-z].*>|!--)`, "i")
  };
  var newline = /^(?:[ \t]*(?:\n|$))+/;
  var blockCode = /^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+/;
  var fences = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}\1[~`]* *(?=\n|$)|$)/;
  var hr = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/;
  var heading = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/;
  var bullet = /(?:[*+-]|\d{1,9}[.)])/;
  var lheadingCore = /^(?!bull |blockCode|fences|blockquote|heading|html|table)((?:.|\n(?!\s*?\n|bull |blockCode|fences|blockquote|heading|html|table))+?)\n {0,3}(=+|-+) *(?:\n+|$)/;
  var lheading = edit(lheadingCore).replace(/bull/g, bullet).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/\|table/g, "").getRegex();
  var lheadingGfm = edit(lheadingCore).replace(/bull/g, bullet).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/table/g, / {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n/).getRegex();
  var _paragraph = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table| +\n)[^\n]+)*)/;
  var blockText = /^[^\n]+/;
  var _blockLabel = /(?!\s*\])(?:\\.|[^\[\]\\])+/;
  var def = edit(/^ {0,3}\[(label)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(title))? *(?:\n+|$)/).replace("label", _blockLabel).replace("title", /(?:"(?:\\"?|[^"\\])*"|'[^'\n]*(?:\n[^'\n]+)*\n?'|\([^()]*\))/).getRegex();
  var list = edit(/^( {0,3}bull)([ \t][^\n]+?)?(?:\n|$)/).replace(/bull/g, bullet).getRegex();
  var _tag = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul";
  var _comment = /<!--(?:-?>|[\s\S]*?(?:-->|$))/;
  var html = edit(
    "^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n+|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>\\n*|$)|<![A-Z][\\s\\S]*?(?:>\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>\\n*|$)|</?(tag)(?: +|\\n|/?>)[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|<(?!script|pre|style|textarea)([a-z][\\w-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|</(?!script|pre|style|textarea)[a-z][\\w-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$))",
    "i"
  ).replace("comment", _comment).replace("tag", _tag).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex();
  var paragraph = edit(_paragraph).replace("hr", hr).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").replace("list", " {0,3}(?:[*+-]|1[.)]) ").replace("html", "</?(?:tag)(?: +|\\n|/?>)|<(?:script|pre|style|textarea|!--)").replace("tag", _tag).getRegex();
  var blockquote = edit(/^( {0,3}> ?(paragraph|[^\n]*)(?:\n|$))+/).replace("paragraph", paragraph).getRegex();
  var blockNormal = {
    blockquote,
    code: blockCode,
    def,
    fences,
    heading,
    hr,
    html,
    lheading,
    list,
    newline,
    paragraph,
    table: noopTest,
    text: blockText
  };
  var gfmTable = edit(
    "^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)"
  ).replace("hr", hr).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", "(?: {4}| {0,3}	)[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").replace("list", " {0,3}(?:[*+-]|1[.)]) ").replace("html", "</?(?:tag)(?: +|\\n|/?>)|<(?:script|pre|style|textarea|!--)").replace("tag", _tag).getRegex();
  var blockGfm = {
    ...blockNormal,
    lheading: lheadingGfm,
    table: gfmTable,
    paragraph: edit(_paragraph).replace("hr", hr).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", gfmTable).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").replace("list", " {0,3}(?:[*+-]|1[.)]) ").replace("html", "</?(?:tag)(?: +|\\n|/?>)|<(?:script|pre|style|textarea|!--)").replace("tag", _tag).getRegex()
  };
  var blockPedantic = {
    ...blockNormal,
    html: edit(
      `^ *(?:comment *(?:\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1> *(?:\\n{2,}|\\s*$)|<tag(?:"[^"]*"|'[^']*'|\\s[^'"/>\\s]*)*?/?> *(?:\\n{2,}|\\s*$))`
    ).replace("comment", _comment).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
    def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
    heading: /^(#{1,6})(.*)(?:\n+|$)/,
    fences: noopTest,
    // fences not supported
    lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
    paragraph: edit(_paragraph).replace("hr", hr).replace("heading", " *#{1,6} *[^\n]").replace("lheading", lheading).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
  };
  var escape = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/;
  var inlineCode = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/;
  var br = /^( {2,}|\\)\n(?!\s*$)/;
  var inlineText = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/;
  var _punctuation = /[\p{P}\p{S}]/u;
  var _punctuationOrSpace = /[\s\p{P}\p{S}]/u;
  var _notPunctuationOrSpace = /[^\s\p{P}\p{S}]/u;
  var punctuation = edit(/^((?![*_])punctSpace)/, "u").replace(/punctSpace/g, _punctuationOrSpace).getRegex();
  var _punctuationGfmStrongEm = /(?!~)[\p{P}\p{S}]/u;
  var _punctuationOrSpaceGfmStrongEm = /(?!~)[\s\p{P}\p{S}]/u;
  var _notPunctuationOrSpaceGfmStrongEm = /(?:[^\s\p{P}\p{S}]|~)/u;
  var blockSkip = /\[[^[\]]*?\]\((?:\\.|[^\\\(\)]|\((?:\\.|[^\\\(\)])*\))*\)|`[^`]*?`|<[^<>]*?>/g;
  var emStrongLDelimCore = /^(?:\*+(?:((?!\*)punct)|[^\s*]))|^_+(?:((?!_)punct)|([^\s_]))/;
  var emStrongLDelim = edit(emStrongLDelimCore, "u").replace(/punct/g, _punctuation).getRegex();
  var emStrongLDelimGfm = edit(emStrongLDelimCore, "u").replace(/punct/g, _punctuationGfmStrongEm).getRegex();
  var emStrongRDelimAstCore = "^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)punctSpace(\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|notPunctSpace(\\*+)(?=notPunctSpace)";
  var emStrongRDelimAst = edit(emStrongRDelimAstCore, "gu").replace(/notPunctSpace/g, _notPunctuationOrSpace).replace(/punctSpace/g, _punctuationOrSpace).replace(/punct/g, _punctuation).getRegex();
  var emStrongRDelimAstGfm = edit(emStrongRDelimAstCore, "gu").replace(/notPunctSpace/g, _notPunctuationOrSpaceGfmStrongEm).replace(/punctSpace/g, _punctuationOrSpaceGfmStrongEm).replace(/punct/g, _punctuationGfmStrongEm).getRegex();
  var emStrongRDelimUnd = edit(
    "^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)punctSpace(_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)",
    "gu"
  ).replace(/notPunctSpace/g, _notPunctuationOrSpace).replace(/punctSpace/g, _punctuationOrSpace).replace(/punct/g, _punctuation).getRegex();
  var anyPunctuation = edit(/\\(punct)/, "gu").replace(/punct/g, _punctuation).getRegex();
  var autolink = edit(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex();
  var _inlineComment = edit(_comment).replace("(?:-->|$)", "-->").getRegex();
  var tag = edit(
    "^comment|^</[a-zA-Z][\\w:-]*\\s*>|^<[a-zA-Z][\\w-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>"
  ).replace("comment", _inlineComment).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex();
  var _inlineLabel = /(?:\[(?:\\.|[^\[\]\\])*\]|\\.|`[^`]*`|[^\[\]\\`])*?/;
  var link = edit(/^!?\[(label)\]\(\s*(href)(?:(?:[ \t]*(?:\n[ \t]*)?)(title))?\s*\)/).replace("label", _inlineLabel).replace("href", /<(?:\\.|[^\n<>\\])+>|[^ \t\n\x00-\x1f]*/).replace("title", /"(?:\\"?|[^"\\])*"|'(?:\\'?|[^'\\])*'|\((?:\\\)?|[^)\\])*\)/).getRegex();
  var reflink = edit(/^!?\[(label)\]\[(ref)\]/).replace("label", _inlineLabel).replace("ref", _blockLabel).getRegex();
  var nolink = edit(/^!?\[(ref)\](?:\[\])?/).replace("ref", _blockLabel).getRegex();
  var reflinkSearch = edit("reflink|nolink(?!\\()", "g").replace("reflink", reflink).replace("nolink", nolink).getRegex();
  var inlineNormal = {
    _backpedal: noopTest,
    // only used for GFM url
    anyPunctuation,
    autolink,
    blockSkip,
    br,
    code: inlineCode,
    del: noopTest,
    emStrongLDelim,
    emStrongRDelimAst,
    emStrongRDelimUnd,
    escape,
    link,
    nolink,
    punctuation,
    reflink,
    reflinkSearch,
    tag,
    text: inlineText,
    url: noopTest
  };
  var inlinePedantic = {
    ...inlineNormal,
    link: edit(/^!?\[(label)\]\((.*?)\)/).replace("label", _inlineLabel).getRegex(),
    reflink: edit(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", _inlineLabel).getRegex()
  };
  var inlineGfm = {
    ...inlineNormal,
    emStrongRDelimAst: emStrongRDelimAstGfm,
    emStrongLDelim: emStrongLDelimGfm,
    url: edit(/^((?:ftp|https?):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/, "i").replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![-_])/).getRegex(),
    _backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
    del: /^(~~?)(?=[^\s~])((?:\\.|[^\\])*?(?:\\.|[^\s~\\]))\1(?=[^~]|$)/,
    text: /^([`~]+|[^`~])(?:(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|https?:\/\/|ftp:\/\/|www\.|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)))/
  };
  var inlineBreaks = {
    ...inlineGfm,
    br: edit(br).replace("{2,}", "*").getRegex(),
    text: edit(inlineGfm.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
  };
  var block = {
    normal: blockNormal,
    gfm: blockGfm,
    pedantic: blockPedantic
  };
  var inline = {
    normal: inlineNormal,
    gfm: inlineGfm,
    breaks: inlineBreaks,
    pedantic: inlinePedantic
  };
  var escapeReplacements = {
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#39;"
  };
  var getEscapeReplacement = (ch) => escapeReplacements[ch];
  function escape2(html22, encode) {
    if (encode) {
      if (other.escapeTest.test(html22)) {
        return html22.replace(other.escapeReplace, getEscapeReplacement);
      }
    } else {
      if (other.escapeTestNoEncode.test(html22)) {
        return html22.replace(other.escapeReplaceNoEncode, getEscapeReplacement);
      }
    }
    return html22;
  }
  function cleanUrl(href) {
    try {
      href = encodeURI(href).replace(other.percentDecode, "%");
    } catch {
      return null;
    }
    return href;
  }
  function splitCells(tableRow, count) {
    const row = tableRow.replace(other.findPipe, (match, offset, str2) => {
      let escaped = false;
      let curr = offset;
      while (--curr >= 0 && str2[curr] === "\\") escaped = !escaped;
      if (escaped) {
        return "|";
      } else {
        return " |";
      }
    }), cells = row.split(other.splitPipe);
    let i = 0;
    if (!cells[0].trim()) {
      cells.shift();
    }
    if (cells.length > 0 && !cells.at(-1)?.trim()) {
      cells.pop();
    }
    if (count) {
      if (cells.length > count) {
        cells.splice(count);
      } else {
        while (cells.length < count) cells.push("");
      }
    }
    for (; i < cells.length; i++) {
      cells[i] = cells[i].trim().replace(other.slashPipe, "|");
    }
    return cells;
  }
  function rtrim(str2, c, invert) {
    const l = str2.length;
    if (l === 0) {
      return "";
    }
    let suffLen = 0;
    while (suffLen < l) {
      const currChar = str2.charAt(l - suffLen - 1);
      if (currChar === c && !invert) {
        suffLen++;
      } else if (currChar !== c && invert) {
        suffLen++;
      } else {
        break;
      }
    }
    return str2.slice(0, l - suffLen);
  }
  function findClosingBracket(str2, b) {
    if (str2.indexOf(b[1]) === -1) {
      return -1;
    }
    let level = 0;
    for (let i = 0; i < str2.length; i++) {
      if (str2[i] === "\\") {
        i++;
      } else if (str2[i] === b[0]) {
        level++;
      } else if (str2[i] === b[1]) {
        level--;
        if (level < 0) {
          return i;
        }
      }
    }
    if (level > 0) {
      return -2;
    }
    return -1;
  }
  function outputLink(cap, link2, raw, lexer2, rules) {
    const href = link2.href;
    const title = link2.title || null;
    const text2 = cap[1].replace(rules.other.outputLinkReplace, "$1");
    lexer2.state.inLink = true;
    const token = {
      type: cap[0].charAt(0) === "!" ? "image" : "link",
      raw,
      href,
      title,
      text: text2,
      tokens: lexer2.inlineTokens(text2)
    };
    lexer2.state.inLink = false;
    return token;
  }
  function indentCodeCompensation(raw, text2, rules) {
    const matchIndentToCode = raw.match(rules.other.indentCodeCompensation);
    if (matchIndentToCode === null) {
      return text2;
    }
    const indentToCode = matchIndentToCode[1];
    return text2.split("\n").map((node) => {
      const matchIndentInNode = node.match(rules.other.beginningSpace);
      if (matchIndentInNode === null) {
        return node;
      }
      const [indentInNode] = matchIndentInNode;
      if (indentInNode.length >= indentToCode.length) {
        return node.slice(indentToCode.length);
      }
      return node;
    }).join("\n");
  }
  var _Tokenizer = class {
    // set by the lexer
    constructor(options2) {
      __publicField(this, "options");
      __publicField(this, "rules");
      // set by the lexer
      __publicField(this, "lexer");
      this.options = options2 || _defaults;
    }
    space(src) {
      const cap = this.rules.block.newline.exec(src);
      if (cap && cap[0].length > 0) {
        return {
          type: "space",
          raw: cap[0]
        };
      }
    }
    code(src) {
      const cap = this.rules.block.code.exec(src);
      if (cap) {
        const text2 = cap[0].replace(this.rules.other.codeRemoveIndent, "");
        return {
          type: "code",
          raw: cap[0],
          codeBlockStyle: "indented",
          text: !this.options.pedantic ? rtrim(text2, "\n") : text2
        };
      }
    }
    fences(src) {
      const cap = this.rules.block.fences.exec(src);
      if (cap) {
        const raw = cap[0];
        const text2 = indentCodeCompensation(raw, cap[3] || "", this.rules);
        return {
          type: "code",
          raw,
          lang: cap[2] ? cap[2].trim().replace(this.rules.inline.anyPunctuation, "$1") : cap[2],
          text: text2
        };
      }
    }
    heading(src) {
      const cap = this.rules.block.heading.exec(src);
      if (cap) {
        let text2 = cap[2].trim();
        if (this.rules.other.endingHash.test(text2)) {
          const trimmed = rtrim(text2, "#");
          if (this.options.pedantic) {
            text2 = trimmed.trim();
          } else if (!trimmed || this.rules.other.endingSpaceChar.test(trimmed)) {
            text2 = trimmed.trim();
          }
        }
        return {
          type: "heading",
          raw: cap[0],
          depth: cap[1].length,
          text: text2,
          tokens: this.lexer.inline(text2)
        };
      }
    }
    hr(src) {
      const cap = this.rules.block.hr.exec(src);
      if (cap) {
        return {
          type: "hr",
          raw: rtrim(cap[0], "\n")
        };
      }
    }
    blockquote(src) {
      const cap = this.rules.block.blockquote.exec(src);
      if (cap) {
        let lines = rtrim(cap[0], "\n").split("\n");
        let raw = "";
        let text2 = "";
        const tokens = [];
        while (lines.length > 0) {
          let inBlockquote = false;
          const currentLines = [];
          let i;
          for (i = 0; i < lines.length; i++) {
            if (this.rules.other.blockquoteStart.test(lines[i])) {
              currentLines.push(lines[i]);
              inBlockquote = true;
            } else if (!inBlockquote) {
              currentLines.push(lines[i]);
            } else {
              break;
            }
          }
          lines = lines.slice(i);
          const currentRaw = currentLines.join("\n");
          const currentText = currentRaw.replace(this.rules.other.blockquoteSetextReplace, "\n    $1").replace(this.rules.other.blockquoteSetextReplace2, "");
          raw = raw ? `${raw}
${currentRaw}` : currentRaw;
          text2 = text2 ? `${text2}
${currentText}` : currentText;
          const top = this.lexer.state.top;
          this.lexer.state.top = true;
          this.lexer.blockTokens(currentText, tokens, true);
          this.lexer.state.top = top;
          if (lines.length === 0) {
            break;
          }
          const lastToken = tokens.at(-1);
          if (lastToken?.type === "code") {
            break;
          } else if (lastToken?.type === "blockquote") {
            const oldToken = lastToken;
            const newText = oldToken.raw + "\n" + lines.join("\n");
            const newToken = this.blockquote(newText);
            tokens[tokens.length - 1] = newToken;
            raw = raw.substring(0, raw.length - oldToken.raw.length) + newToken.raw;
            text2 = text2.substring(0, text2.length - oldToken.text.length) + newToken.text;
            break;
          } else if (lastToken?.type === "list") {
            const oldToken = lastToken;
            const newText = oldToken.raw + "\n" + lines.join("\n");
            const newToken = this.list(newText);
            tokens[tokens.length - 1] = newToken;
            raw = raw.substring(0, raw.length - lastToken.raw.length) + newToken.raw;
            text2 = text2.substring(0, text2.length - oldToken.raw.length) + newToken.raw;
            lines = newText.substring(tokens.at(-1).raw.length).split("\n");
            continue;
          }
        }
        return {
          type: "blockquote",
          raw,
          tokens,
          text: text2
        };
      }
    }
    list(src) {
      let cap = this.rules.block.list.exec(src);
      if (cap) {
        let bull = cap[1].trim();
        const isordered = bull.length > 1;
        const list2 = {
          type: "list",
          raw: "",
          ordered: isordered,
          start: isordered ? +bull.slice(0, -1) : "",
          loose: false,
          items: []
        };
        bull = isordered ? `\\d{1,9}\\${bull.slice(-1)}` : `\\${bull}`;
        if (this.options.pedantic) {
          bull = isordered ? bull : "[*+-]";
        }
        const itemRegex = this.rules.other.listItemRegex(bull);
        let endsWithBlankLine = false;
        while (src) {
          let endEarly = false;
          let raw = "";
          let itemContents = "";
          if (!(cap = itemRegex.exec(src))) {
            break;
          }
          if (this.rules.block.hr.test(src)) {
            break;
          }
          raw = cap[0];
          src = src.substring(raw.length);
          let line = cap[2].split("\n", 1)[0].replace(this.rules.other.listReplaceTabs, (t2) => " ".repeat(3 * t2.length));
          let nextLine = src.split("\n", 1)[0];
          let blankLine = !line.trim();
          let indent = 0;
          if (this.options.pedantic) {
            indent = 2;
            itemContents = line.trimStart();
          } else if (blankLine) {
            indent = cap[1].length + 1;
          } else {
            indent = cap[2].search(this.rules.other.nonSpaceChar);
            indent = indent > 4 ? 1 : indent;
            itemContents = line.slice(indent);
            indent += cap[1].length;
          }
          if (blankLine && this.rules.other.blankLine.test(nextLine)) {
            raw += nextLine + "\n";
            src = src.substring(nextLine.length + 1);
            endEarly = true;
          }
          if (!endEarly) {
            const nextBulletRegex = this.rules.other.nextBulletRegex(indent);
            const hrRegex = this.rules.other.hrRegex(indent);
            const fencesBeginRegex = this.rules.other.fencesBeginRegex(indent);
            const headingBeginRegex = this.rules.other.headingBeginRegex(indent);
            const htmlBeginRegex = this.rules.other.htmlBeginRegex(indent);
            while (src) {
              const rawLine = src.split("\n", 1)[0];
              let nextLineWithoutTabs;
              nextLine = rawLine;
              if (this.options.pedantic) {
                nextLine = nextLine.replace(this.rules.other.listReplaceNesting, "  ");
                nextLineWithoutTabs = nextLine;
              } else {
                nextLineWithoutTabs = nextLine.replace(this.rules.other.tabCharGlobal, "    ");
              }
              if (fencesBeginRegex.test(nextLine)) {
                break;
              }
              if (headingBeginRegex.test(nextLine)) {
                break;
              }
              if (htmlBeginRegex.test(nextLine)) {
                break;
              }
              if (nextBulletRegex.test(nextLine)) {
                break;
              }
              if (hrRegex.test(nextLine)) {
                break;
              }
              if (nextLineWithoutTabs.search(this.rules.other.nonSpaceChar) >= indent || !nextLine.trim()) {
                itemContents += "\n" + nextLineWithoutTabs.slice(indent);
              } else {
                if (blankLine) {
                  break;
                }
                if (line.replace(this.rules.other.tabCharGlobal, "    ").search(this.rules.other.nonSpaceChar) >= 4) {
                  break;
                }
                if (fencesBeginRegex.test(line)) {
                  break;
                }
                if (headingBeginRegex.test(line)) {
                  break;
                }
                if (hrRegex.test(line)) {
                  break;
                }
                itemContents += "\n" + nextLine;
              }
              if (!blankLine && !nextLine.trim()) {
                blankLine = true;
              }
              raw += rawLine + "\n";
              src = src.substring(rawLine.length + 1);
              line = nextLineWithoutTabs.slice(indent);
            }
          }
          if (!list2.loose) {
            if (endsWithBlankLine) {
              list2.loose = true;
            } else if (this.rules.other.doubleBlankLine.test(raw)) {
              endsWithBlankLine = true;
            }
          }
          let istask = null;
          let ischecked;
          if (this.options.gfm) {
            istask = this.rules.other.listIsTask.exec(itemContents);
            if (istask) {
              ischecked = istask[0] !== "[ ] ";
              itemContents = itemContents.replace(this.rules.other.listReplaceTask, "");
            }
          }
          list2.items.push({
            type: "list_item",
            raw,
            task: !!istask,
            checked: ischecked,
            loose: false,
            text: itemContents,
            tokens: []
          });
          list2.raw += raw;
        }
        const lastItem = list2.items.at(-1);
        if (lastItem) {
          lastItem.raw = lastItem.raw.trimEnd();
          lastItem.text = lastItem.text.trimEnd();
        } else {
          return;
        }
        list2.raw = list2.raw.trimEnd();
        for (let i = 0; i < list2.items.length; i++) {
          this.lexer.state.top = false;
          list2.items[i].tokens = this.lexer.blockTokens(list2.items[i].text, []);
          if (!list2.loose) {
            const spacers = list2.items[i].tokens.filter((t2) => t2.type === "space");
            const hasMultipleLineBreaks = spacers.length > 0 && spacers.some((t2) => this.rules.other.anyLine.test(t2.raw));
            list2.loose = hasMultipleLineBreaks;
          }
        }
        if (list2.loose) {
          for (let i = 0; i < list2.items.length; i++) {
            list2.items[i].loose = true;
          }
        }
        return list2;
      }
    }
    html(src) {
      const cap = this.rules.block.html.exec(src);
      if (cap) {
        const token = {
          type: "html",
          block: true,
          raw: cap[0],
          pre: cap[1] === "pre" || cap[1] === "script" || cap[1] === "style",
          text: cap[0]
        };
        return token;
      }
    }
    def(src) {
      const cap = this.rules.block.def.exec(src);
      if (cap) {
        const tag2 = cap[1].toLowerCase().replace(this.rules.other.multipleSpaceGlobal, " ");
        const href = cap[2] ? cap[2].replace(this.rules.other.hrefBrackets, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "";
        const title = cap[3] ? cap[3].substring(1, cap[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : cap[3];
        return {
          type: "def",
          tag: tag2,
          raw: cap[0],
          href,
          title
        };
      }
    }
    table(src) {
      const cap = this.rules.block.table.exec(src);
      if (!cap) {
        return;
      }
      if (!this.rules.other.tableDelimiter.test(cap[2])) {
        return;
      }
      const headers = splitCells(cap[1]);
      const aligns = cap[2].replace(this.rules.other.tableAlignChars, "").split("|");
      const rows = cap[3]?.trim() ? cap[3].replace(this.rules.other.tableRowBlankLine, "").split("\n") : [];
      const item = {
        type: "table",
        raw: cap[0],
        header: [],
        align: [],
        rows: []
      };
      if (headers.length !== aligns.length) {
        return;
      }
      for (const align of aligns) {
        if (this.rules.other.tableAlignRight.test(align)) {
          item.align.push("right");
        } else if (this.rules.other.tableAlignCenter.test(align)) {
          item.align.push("center");
        } else if (this.rules.other.tableAlignLeft.test(align)) {
          item.align.push("left");
        } else {
          item.align.push(null);
        }
      }
      for (let i = 0; i < headers.length; i++) {
        item.header.push({
          text: headers[i],
          tokens: this.lexer.inline(headers[i]),
          header: true,
          align: item.align[i]
        });
      }
      for (const row of rows) {
        item.rows.push(splitCells(row, item.header.length).map((cell, i) => {
          return {
            text: cell,
            tokens: this.lexer.inline(cell),
            header: false,
            align: item.align[i]
          };
        }));
      }
      return item;
    }
    lheading(src) {
      const cap = this.rules.block.lheading.exec(src);
      if (cap) {
        return {
          type: "heading",
          raw: cap[0],
          depth: cap[2].charAt(0) === "=" ? 1 : 2,
          text: cap[1],
          tokens: this.lexer.inline(cap[1])
        };
      }
    }
    paragraph(src) {
      const cap = this.rules.block.paragraph.exec(src);
      if (cap) {
        const text2 = cap[1].charAt(cap[1].length - 1) === "\n" ? cap[1].slice(0, -1) : cap[1];
        return {
          type: "paragraph",
          raw: cap[0],
          text: text2,
          tokens: this.lexer.inline(text2)
        };
      }
    }
    text(src) {
      const cap = this.rules.block.text.exec(src);
      if (cap) {
        return {
          type: "text",
          raw: cap[0],
          text: cap[0],
          tokens: this.lexer.inline(cap[0])
        };
      }
    }
    escape(src) {
      const cap = this.rules.inline.escape.exec(src);
      if (cap) {
        return {
          type: "escape",
          raw: cap[0],
          text: cap[1]
        };
      }
    }
    tag(src) {
      const cap = this.rules.inline.tag.exec(src);
      if (cap) {
        if (!this.lexer.state.inLink && this.rules.other.startATag.test(cap[0])) {
          this.lexer.state.inLink = true;
        } else if (this.lexer.state.inLink && this.rules.other.endATag.test(cap[0])) {
          this.lexer.state.inLink = false;
        }
        if (!this.lexer.state.inRawBlock && this.rules.other.startPreScriptTag.test(cap[0])) {
          this.lexer.state.inRawBlock = true;
        } else if (this.lexer.state.inRawBlock && this.rules.other.endPreScriptTag.test(cap[0])) {
          this.lexer.state.inRawBlock = false;
        }
        return {
          type: "html",
          raw: cap[0],
          inLink: this.lexer.state.inLink,
          inRawBlock: this.lexer.state.inRawBlock,
          block: false,
          text: cap[0]
        };
      }
    }
    link(src) {
      const cap = this.rules.inline.link.exec(src);
      if (cap) {
        const trimmedUrl = cap[2].trim();
        if (!this.options.pedantic && this.rules.other.startAngleBracket.test(trimmedUrl)) {
          if (!this.rules.other.endAngleBracket.test(trimmedUrl)) {
            return;
          }
          const rtrimSlash = rtrim(trimmedUrl.slice(0, -1), "\\");
          if ((trimmedUrl.length - rtrimSlash.length) % 2 === 0) {
            return;
          }
        } else {
          const lastParenIndex = findClosingBracket(cap[2], "()");
          if (lastParenIndex === -2) {
            return;
          }
          if (lastParenIndex > -1) {
            const start = cap[0].indexOf("!") === 0 ? 5 : 4;
            const linkLen = start + cap[1].length + lastParenIndex;
            cap[2] = cap[2].substring(0, lastParenIndex);
            cap[0] = cap[0].substring(0, linkLen).trim();
            cap[3] = "";
          }
        }
        let href = cap[2];
        let title = "";
        if (this.options.pedantic) {
          const link2 = this.rules.other.pedanticHrefTitle.exec(href);
          if (link2) {
            href = link2[1];
            title = link2[3];
          }
        } else {
          title = cap[3] ? cap[3].slice(1, -1) : "";
        }
        href = href.trim();
        if (this.rules.other.startAngleBracket.test(href)) {
          if (this.options.pedantic && !this.rules.other.endAngleBracket.test(trimmedUrl)) {
            href = href.slice(1);
          } else {
            href = href.slice(1, -1);
          }
        }
        return outputLink(cap, {
          href: href ? href.replace(this.rules.inline.anyPunctuation, "$1") : href,
          title: title ? title.replace(this.rules.inline.anyPunctuation, "$1") : title
        }, cap[0], this.lexer, this.rules);
      }
    }
    reflink(src, links) {
      let cap;
      if ((cap = this.rules.inline.reflink.exec(src)) || (cap = this.rules.inline.nolink.exec(src))) {
        const linkString = (cap[2] || cap[1]).replace(this.rules.other.multipleSpaceGlobal, " ");
        const link2 = links[linkString.toLowerCase()];
        if (!link2) {
          const text2 = cap[0].charAt(0);
          return {
            type: "text",
            raw: text2,
            text: text2
          };
        }
        return outputLink(cap, link2, cap[0], this.lexer, this.rules);
      }
    }
    emStrong(src, maskedSrc, prevChar = "") {
      let match = this.rules.inline.emStrongLDelim.exec(src);
      if (!match) return;
      if (match[3] && prevChar.match(this.rules.other.unicodeAlphaNumeric)) return;
      const nextChar = match[1] || match[2] || "";
      if (!nextChar || !prevChar || this.rules.inline.punctuation.exec(prevChar)) {
        const lLength = [...match[0]].length - 1;
        let rDelim, rLength, delimTotal = lLength, midDelimTotal = 0;
        const endReg = match[0][0] === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd;
        endReg.lastIndex = 0;
        maskedSrc = maskedSrc.slice(-1 * src.length + lLength);
        while ((match = endReg.exec(maskedSrc)) != null) {
          rDelim = match[1] || match[2] || match[3] || match[4] || match[5] || match[6];
          if (!rDelim) continue;
          rLength = [...rDelim].length;
          if (match[3] || match[4]) {
            delimTotal += rLength;
            continue;
          } else if (match[5] || match[6]) {
            if (lLength % 3 && !((lLength + rLength) % 3)) {
              midDelimTotal += rLength;
              continue;
            }
          }
          delimTotal -= rLength;
          if (delimTotal > 0) continue;
          rLength = Math.min(rLength, rLength + delimTotal + midDelimTotal);
          const lastCharLength = [...match[0]][0].length;
          const raw = src.slice(0, lLength + match.index + lastCharLength + rLength);
          if (Math.min(lLength, rLength) % 2) {
            const text22 = raw.slice(1, -1);
            return {
              type: "em",
              raw,
              text: text22,
              tokens: this.lexer.inlineTokens(text22)
            };
          }
          const text2 = raw.slice(2, -2);
          return {
            type: "strong",
            raw,
            text: text2,
            tokens: this.lexer.inlineTokens(text2)
          };
        }
      }
    }
    codespan(src) {
      const cap = this.rules.inline.code.exec(src);
      if (cap) {
        let text2 = cap[2].replace(this.rules.other.newLineCharGlobal, " ");
        const hasNonSpaceChars = this.rules.other.nonSpaceChar.test(text2);
        const hasSpaceCharsOnBothEnds = this.rules.other.startingSpaceChar.test(text2) && this.rules.other.endingSpaceChar.test(text2);
        if (hasNonSpaceChars && hasSpaceCharsOnBothEnds) {
          text2 = text2.substring(1, text2.length - 1);
        }
        return {
          type: "codespan",
          raw: cap[0],
          text: text2
        };
      }
    }
    br(src) {
      const cap = this.rules.inline.br.exec(src);
      if (cap) {
        return {
          type: "br",
          raw: cap[0]
        };
      }
    }
    del(src) {
      const cap = this.rules.inline.del.exec(src);
      if (cap) {
        return {
          type: "del",
          raw: cap[0],
          text: cap[2],
          tokens: this.lexer.inlineTokens(cap[2])
        };
      }
    }
    autolink(src) {
      const cap = this.rules.inline.autolink.exec(src);
      if (cap) {
        let text2, href;
        if (cap[2] === "@") {
          text2 = cap[1];
          href = "mailto:" + text2;
        } else {
          text2 = cap[1];
          href = text2;
        }
        return {
          type: "link",
          raw: cap[0],
          text: text2,
          href,
          tokens: [
            {
              type: "text",
              raw: text2,
              text: text2
            }
          ]
        };
      }
    }
    url(src) {
      let cap;
      if (cap = this.rules.inline.url.exec(src)) {
        let text2, href;
        if (cap[2] === "@") {
          text2 = cap[0];
          href = "mailto:" + text2;
        } else {
          let prevCapZero;
          do {
            prevCapZero = cap[0];
            cap[0] = this.rules.inline._backpedal.exec(cap[0])?.[0] ?? "";
          } while (prevCapZero !== cap[0]);
          text2 = cap[0];
          if (cap[1] === "www.") {
            href = "http://" + cap[0];
          } else {
            href = cap[0];
          }
        }
        return {
          type: "link",
          raw: cap[0],
          text: text2,
          href,
          tokens: [
            {
              type: "text",
              raw: text2,
              text: text2
            }
          ]
        };
      }
    }
    inlineText(src) {
      const cap = this.rules.inline.text.exec(src);
      if (cap) {
        const escaped = this.lexer.state.inRawBlock;
        return {
          type: "text",
          raw: cap[0],
          text: cap[0],
          escaped
        };
      }
    }
  };
  var _Lexer = class __Lexer {
    constructor(options2) {
      __publicField(this, "tokens");
      __publicField(this, "options");
      __publicField(this, "state");
      __publicField(this, "tokenizer");
      __publicField(this, "inlineQueue");
      this.tokens = [];
      this.tokens.links = /* @__PURE__ */ Object.create(null);
      this.options = options2 || _defaults;
      this.options.tokenizer = this.options.tokenizer || new _Tokenizer();
      this.tokenizer = this.options.tokenizer;
      this.tokenizer.options = this.options;
      this.tokenizer.lexer = this;
      this.inlineQueue = [];
      this.state = {
        inLink: false,
        inRawBlock: false,
        top: true
      };
      const rules = {
        other,
        block: block.normal,
        inline: inline.normal
      };
      if (this.options.pedantic) {
        rules.block = block.pedantic;
        rules.inline = inline.pedantic;
      } else if (this.options.gfm) {
        rules.block = block.gfm;
        if (this.options.breaks) {
          rules.inline = inline.breaks;
        } else {
          rules.inline = inline.gfm;
        }
      }
      this.tokenizer.rules = rules;
    }
    /**
     * Expose Rules
     */
    static get rules() {
      return {
        block,
        inline
      };
    }
    /**
     * Static Lex Method
     */
    static lex(src, options2) {
      const lexer2 = new __Lexer(options2);
      return lexer2.lex(src);
    }
    /**
     * Static Lex Inline Method
     */
    static lexInline(src, options2) {
      const lexer2 = new __Lexer(options2);
      return lexer2.inlineTokens(src);
    }
    /**
     * Preprocessing
     */
    lex(src) {
      src = src.replace(other.carriageReturn, "\n");
      this.blockTokens(src, this.tokens);
      for (let i = 0; i < this.inlineQueue.length; i++) {
        const next = this.inlineQueue[i];
        this.inlineTokens(next.src, next.tokens);
      }
      this.inlineQueue = [];
      return this.tokens;
    }
    blockTokens(src, tokens = [], lastParagraphClipped = false) {
      if (this.options.pedantic) {
        src = src.replace(other.tabCharGlobal, "    ").replace(other.spaceLine, "");
      }
      while (src) {
        let token;
        if (this.options.extensions?.block?.some((extTokenizer) => {
          if (token = extTokenizer.call({ lexer: this }, src, tokens)) {
            src = src.substring(token.raw.length);
            tokens.push(token);
            return true;
          }
          return false;
        })) {
          continue;
        }
        if (token = this.tokenizer.space(src)) {
          src = src.substring(token.raw.length);
          const lastToken = tokens.at(-1);
          if (token.raw.length === 1 && lastToken !== void 0) {
            lastToken.raw += "\n";
          } else {
            tokens.push(token);
          }
          continue;
        }
        if (token = this.tokenizer.code(src)) {
          src = src.substring(token.raw.length);
          const lastToken = tokens.at(-1);
          if (lastToken?.type === "paragraph" || lastToken?.type === "text") {
            lastToken.raw += "\n" + token.raw;
            lastToken.text += "\n" + token.text;
            this.inlineQueue.at(-1).src = lastToken.text;
          } else {
            tokens.push(token);
          }
          continue;
        }
        if (token = this.tokenizer.fences(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.heading(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.hr(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.blockquote(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.list(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.html(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.def(src)) {
          src = src.substring(token.raw.length);
          const lastToken = tokens.at(-1);
          if (lastToken?.type === "paragraph" || lastToken?.type === "text") {
            lastToken.raw += "\n" + token.raw;
            lastToken.text += "\n" + token.raw;
            this.inlineQueue.at(-1).src = lastToken.text;
          } else if (!this.tokens.links[token.tag]) {
            this.tokens.links[token.tag] = {
              href: token.href,
              title: token.title
            };
          }
          continue;
        }
        if (token = this.tokenizer.table(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.lheading(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        let cutSrc = src;
        if (this.options.extensions?.startBlock) {
          let startIndex = Infinity;
          const tempSrc = src.slice(1);
          let tempStart;
          this.options.extensions.startBlock.forEach((getStartIndex) => {
            tempStart = getStartIndex.call({ lexer: this }, tempSrc);
            if (typeof tempStart === "number" && tempStart >= 0) {
              startIndex = Math.min(startIndex, tempStart);
            }
          });
          if (startIndex < Infinity && startIndex >= 0) {
            cutSrc = src.substring(0, startIndex + 1);
          }
        }
        if (this.state.top && (token = this.tokenizer.paragraph(cutSrc))) {
          const lastToken = tokens.at(-1);
          if (lastParagraphClipped && lastToken?.type === "paragraph") {
            lastToken.raw += "\n" + token.raw;
            lastToken.text += "\n" + token.text;
            this.inlineQueue.pop();
            this.inlineQueue.at(-1).src = lastToken.text;
          } else {
            tokens.push(token);
          }
          lastParagraphClipped = cutSrc.length !== src.length;
          src = src.substring(token.raw.length);
          continue;
        }
        if (token = this.tokenizer.text(src)) {
          src = src.substring(token.raw.length);
          const lastToken = tokens.at(-1);
          if (lastToken?.type === "text") {
            lastToken.raw += "\n" + token.raw;
            lastToken.text += "\n" + token.text;
            this.inlineQueue.pop();
            this.inlineQueue.at(-1).src = lastToken.text;
          } else {
            tokens.push(token);
          }
          continue;
        }
        if (src) {
          const errMsg = "Infinite loop on byte: " + src.charCodeAt(0);
          if (this.options.silent) {
            console.error(errMsg);
            break;
          } else {
            throw new Error(errMsg);
          }
        }
      }
      this.state.top = true;
      return tokens;
    }
    inline(src, tokens = []) {
      this.inlineQueue.push({ src, tokens });
      return tokens;
    }
    /**
     * Lexing/Compiling
     */
    inlineTokens(src, tokens = []) {
      let maskedSrc = src;
      let match = null;
      if (this.tokens.links) {
        const links = Object.keys(this.tokens.links);
        if (links.length > 0) {
          while ((match = this.tokenizer.rules.inline.reflinkSearch.exec(maskedSrc)) != null) {
            if (links.includes(match[0].slice(match[0].lastIndexOf("[") + 1, -1))) {
              maskedSrc = maskedSrc.slice(0, match.index) + "[" + "a".repeat(match[0].length - 2) + "]" + maskedSrc.slice(this.tokenizer.rules.inline.reflinkSearch.lastIndex);
            }
          }
        }
      }
      while ((match = this.tokenizer.rules.inline.anyPunctuation.exec(maskedSrc)) != null) {
        maskedSrc = maskedSrc.slice(0, match.index) + "++" + maskedSrc.slice(this.tokenizer.rules.inline.anyPunctuation.lastIndex);
      }
      while ((match = this.tokenizer.rules.inline.blockSkip.exec(maskedSrc)) != null) {
        maskedSrc = maskedSrc.slice(0, match.index) + "[" + "a".repeat(match[0].length - 2) + "]" + maskedSrc.slice(this.tokenizer.rules.inline.blockSkip.lastIndex);
      }
      let keepPrevChar = false;
      let prevChar = "";
      while (src) {
        if (!keepPrevChar) {
          prevChar = "";
        }
        keepPrevChar = false;
        let token;
        if (this.options.extensions?.inline?.some((extTokenizer) => {
          if (token = extTokenizer.call({ lexer: this }, src, tokens)) {
            src = src.substring(token.raw.length);
            tokens.push(token);
            return true;
          }
          return false;
        })) {
          continue;
        }
        if (token = this.tokenizer.escape(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.tag(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.link(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.reflink(src, this.tokens.links)) {
          src = src.substring(token.raw.length);
          const lastToken = tokens.at(-1);
          if (token.type === "text" && lastToken?.type === "text") {
            lastToken.raw += token.raw;
            lastToken.text += token.text;
          } else {
            tokens.push(token);
          }
          continue;
        }
        if (token = this.tokenizer.emStrong(src, maskedSrc, prevChar)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.codespan(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.br(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.del(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.autolink(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (!this.state.inLink && (token = this.tokenizer.url(src))) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        let cutSrc = src;
        if (this.options.extensions?.startInline) {
          let startIndex = Infinity;
          const tempSrc = src.slice(1);
          let tempStart;
          this.options.extensions.startInline.forEach((getStartIndex) => {
            tempStart = getStartIndex.call({ lexer: this }, tempSrc);
            if (typeof tempStart === "number" && tempStart >= 0) {
              startIndex = Math.min(startIndex, tempStart);
            }
          });
          if (startIndex < Infinity && startIndex >= 0) {
            cutSrc = src.substring(0, startIndex + 1);
          }
        }
        if (token = this.tokenizer.inlineText(cutSrc)) {
          src = src.substring(token.raw.length);
          if (token.raw.slice(-1) !== "_") {
            prevChar = token.raw.slice(-1);
          }
          keepPrevChar = true;
          const lastToken = tokens.at(-1);
          if (lastToken?.type === "text") {
            lastToken.raw += token.raw;
            lastToken.text += token.text;
          } else {
            tokens.push(token);
          }
          continue;
        }
        if (src) {
          const errMsg = "Infinite loop on byte: " + src.charCodeAt(0);
          if (this.options.silent) {
            console.error(errMsg);
            break;
          } else {
            throw new Error(errMsg);
          }
        }
      }
      return tokens;
    }
  };
  var _Renderer = class {
    // set by the parser
    constructor(options2) {
      __publicField(this, "options");
      __publicField(this, "parser");
      this.options = options2 || _defaults;
    }
    space(token) {
      return "";
    }
    code({ text: text2, lang, escaped }) {
      const langString = (lang || "").match(other.notSpaceStart)?.[0];
      const code = text2.replace(other.endingNewline, "") + "\n";
      if (!langString) {
        return "<pre><code>" + (escaped ? code : escape2(code, true)) + "</code></pre>\n";
      }
      return '<pre><code class="language-' + escape2(langString) + '">' + (escaped ? code : escape2(code, true)) + "</code></pre>\n";
    }
    blockquote({ tokens }) {
      const body = this.parser.parse(tokens);
      return `<blockquote>
${body}</blockquote>
`;
    }
    html({ text: text2 }) {
      return text2;
    }
    heading({ tokens, depth }) {
      return `<h${depth}>${this.parser.parseInline(tokens)}</h${depth}>
`;
    }
    hr(token) {
      return "<hr>\n";
    }
    list(token) {
      const ordered = token.ordered;
      const start = token.start;
      let body = "";
      for (let j = 0; j < token.items.length; j++) {
        const item = token.items[j];
        body += this.listitem(item);
      }
      const type2 = ordered ? "ol" : "ul";
      const startAttr = ordered && start !== 1 ? ' start="' + start + '"' : "";
      return "<" + type2 + startAttr + ">\n" + body + "</" + type2 + ">\n";
    }
    listitem(item) {
      let itemBody = "";
      if (item.task) {
        const checkbox = this.checkbox({ checked: !!item.checked });
        if (item.loose) {
          if (item.tokens[0]?.type === "paragraph") {
            item.tokens[0].text = checkbox + " " + item.tokens[0].text;
            if (item.tokens[0].tokens && item.tokens[0].tokens.length > 0 && item.tokens[0].tokens[0].type === "text") {
              item.tokens[0].tokens[0].text = checkbox + " " + escape2(item.tokens[0].tokens[0].text);
              item.tokens[0].tokens[0].escaped = true;
            }
          } else {
            item.tokens.unshift({
              type: "text",
              raw: checkbox + " ",
              text: checkbox + " ",
              escaped: true
            });
          }
        } else {
          itemBody += checkbox + " ";
        }
      }
      itemBody += this.parser.parse(item.tokens, !!item.loose);
      return `<li>${itemBody}</li>
`;
    }
    checkbox({ checked }) {
      return "<input " + (checked ? 'checked="" ' : "") + 'disabled="" type="checkbox">';
    }
    paragraph({ tokens }) {
      return `<p>${this.parser.parseInline(tokens)}</p>
`;
    }
    table(token) {
      let header = "";
      let cell = "";
      for (let j = 0; j < token.header.length; j++) {
        cell += this.tablecell(token.header[j]);
      }
      header += this.tablerow({ text: cell });
      let body = "";
      for (let j = 0; j < token.rows.length; j++) {
        const row = token.rows[j];
        cell = "";
        for (let k = 0; k < row.length; k++) {
          cell += this.tablecell(row[k]);
        }
        body += this.tablerow({ text: cell });
      }
      if (body) body = `<tbody>${body}</tbody>`;
      return "<table>\n<thead>\n" + header + "</thead>\n" + body + "</table>\n";
    }
    tablerow({ text: text2 }) {
      return `<tr>
${text2}</tr>
`;
    }
    tablecell(token) {
      const content = this.parser.parseInline(token.tokens);
      const type2 = token.header ? "th" : "td";
      const tag2 = token.align ? `<${type2} align="${token.align}">` : `<${type2}>`;
      return tag2 + content + `</${type2}>
`;
    }
    /**
     * span level renderer
     */
    strong({ tokens }) {
      return `<strong>${this.parser.parseInline(tokens)}</strong>`;
    }
    em({ tokens }) {
      return `<em>${this.parser.parseInline(tokens)}</em>`;
    }
    codespan({ text: text2 }) {
      return `<code>${escape2(text2, true)}</code>`;
    }
    br(token) {
      return "<br>";
    }
    del({ tokens }) {
      return `<del>${this.parser.parseInline(tokens)}</del>`;
    }
    link({ href, title, tokens }) {
      const text2 = this.parser.parseInline(tokens);
      const cleanHref = cleanUrl(href);
      if (cleanHref === null) {
        return text2;
      }
      href = cleanHref;
      let out = '<a href="' + href + '"';
      if (title) {
        out += ' title="' + escape2(title) + '"';
      }
      out += ">" + text2 + "</a>";
      return out;
    }
    image({ href, title, text: text2, tokens }) {
      if (tokens) {
        text2 = this.parser.parseInline(tokens, this.parser.textRenderer);
      }
      const cleanHref = cleanUrl(href);
      if (cleanHref === null) {
        return escape2(text2);
      }
      href = cleanHref;
      let out = `<img src="${href}" alt="${text2}"`;
      if (title) {
        out += ` title="${escape2(title)}"`;
      }
      out += ">";
      return out;
    }
    text(token) {
      return "tokens" in token && token.tokens ? this.parser.parseInline(token.tokens) : "escaped" in token && token.escaped ? token.text : escape2(token.text);
    }
  };
  var _TextRenderer = class {
    // no need for block level renderers
    strong({ text: text2 }) {
      return text2;
    }
    em({ text: text2 }) {
      return text2;
    }
    codespan({ text: text2 }) {
      return text2;
    }
    del({ text: text2 }) {
      return text2;
    }
    html({ text: text2 }) {
      return text2;
    }
    text({ text: text2 }) {
      return text2;
    }
    link({ text: text2 }) {
      return "" + text2;
    }
    image({ text: text2 }) {
      return "" + text2;
    }
    br() {
      return "";
    }
  };
  var _Parser = class __Parser {
    constructor(options2) {
      __publicField(this, "options");
      __publicField(this, "renderer");
      __publicField(this, "textRenderer");
      this.options = options2 || _defaults;
      this.options.renderer = this.options.renderer || new _Renderer();
      this.renderer = this.options.renderer;
      this.renderer.options = this.options;
      this.renderer.parser = this;
      this.textRenderer = new _TextRenderer();
    }
    /**
     * Static Parse Method
     */
    static parse(tokens, options2) {
      const parser2 = new __Parser(options2);
      return parser2.parse(tokens);
    }
    /**
     * Static Parse Inline Method
     */
    static parseInline(tokens, options2) {
      const parser2 = new __Parser(options2);
      return parser2.parseInline(tokens);
    }
    /**
     * Parse Loop
     */
    parse(tokens, top = true) {
      let out = "";
      for (let i = 0; i < tokens.length; i++) {
        const anyToken = tokens[i];
        if (this.options.extensions?.renderers?.[anyToken.type]) {
          const genericToken = anyToken;
          const ret = this.options.extensions.renderers[genericToken.type].call({ parser: this }, genericToken);
          if (ret !== false || !["space", "hr", "heading", "code", "table", "blockquote", "list", "html", "paragraph", "text"].includes(genericToken.type)) {
            out += ret || "";
            continue;
          }
        }
        const token = anyToken;
        switch (token.type) {
          case "space": {
            out += this.renderer.space(token);
            continue;
          }
          case "hr": {
            out += this.renderer.hr(token);
            continue;
          }
          case "heading": {
            out += this.renderer.heading(token);
            continue;
          }
          case "code": {
            out += this.renderer.code(token);
            continue;
          }
          case "table": {
            out += this.renderer.table(token);
            continue;
          }
          case "blockquote": {
            out += this.renderer.blockquote(token);
            continue;
          }
          case "list": {
            out += this.renderer.list(token);
            continue;
          }
          case "html": {
            out += this.renderer.html(token);
            continue;
          }
          case "paragraph": {
            out += this.renderer.paragraph(token);
            continue;
          }
          case "text": {
            let textToken = token;
            let body = this.renderer.text(textToken);
            while (i + 1 < tokens.length && tokens[i + 1].type === "text") {
              textToken = tokens[++i];
              body += "\n" + this.renderer.text(textToken);
            }
            if (top) {
              out += this.renderer.paragraph({
                type: "paragraph",
                raw: body,
                text: body,
                tokens: [{ type: "text", raw: body, text: body, escaped: true }]
              });
            } else {
              out += body;
            }
            continue;
          }
          default: {
            const errMsg = 'Token with "' + token.type + '" type was not found.';
            if (this.options.silent) {
              console.error(errMsg);
              return "";
            } else {
              throw new Error(errMsg);
            }
          }
        }
      }
      return out;
    }
    /**
     * Parse Inline Tokens
     */
    parseInline(tokens, renderer = this.renderer) {
      let out = "";
      for (let i = 0; i < tokens.length; i++) {
        const anyToken = tokens[i];
        if (this.options.extensions?.renderers?.[anyToken.type]) {
          const ret = this.options.extensions.renderers[anyToken.type].call({ parser: this }, anyToken);
          if (ret !== false || !["escape", "html", "link", "image", "strong", "em", "codespan", "br", "del", "text"].includes(anyToken.type)) {
            out += ret || "";
            continue;
          }
        }
        const token = anyToken;
        switch (token.type) {
          case "escape": {
            out += renderer.text(token);
            break;
          }
          case "html": {
            out += renderer.html(token);
            break;
          }
          case "link": {
            out += renderer.link(token);
            break;
          }
          case "image": {
            out += renderer.image(token);
            break;
          }
          case "strong": {
            out += renderer.strong(token);
            break;
          }
          case "em": {
            out += renderer.em(token);
            break;
          }
          case "codespan": {
            out += renderer.codespan(token);
            break;
          }
          case "br": {
            out += renderer.br(token);
            break;
          }
          case "del": {
            out += renderer.del(token);
            break;
          }
          case "text": {
            out += renderer.text(token);
            break;
          }
          default: {
            const errMsg = 'Token with "' + token.type + '" type was not found.';
            if (this.options.silent) {
              console.error(errMsg);
              return "";
            } else {
              throw new Error(errMsg);
            }
          }
        }
      }
      return out;
    }
  };
  var _a;
  var _Hooks = (_a = class {
    constructor(options2) {
      __publicField(this, "options");
      __publicField(this, "block");
      this.options = options2 || _defaults;
    }
    /**
     * Process markdown before marked
     */
    preprocess(markdown) {
      return markdown;
    }
    /**
     * Process HTML after marked is finished
     */
    postprocess(html22) {
      return html22;
    }
    /**
     * Process all tokens before walk tokens
     */
    processAllTokens(tokens) {
      return tokens;
    }
    /**
     * Provide function to tokenize markdown
     */
    provideLexer() {
      return this.block ? _Lexer.lex : _Lexer.lexInline;
    }
    /**
     * Provide function to parse tokens
     */
    provideParser() {
      return this.block ? _Parser.parse : _Parser.parseInline;
    }
  }, __publicField(_a, "passThroughHooks", /* @__PURE__ */ new Set([
    "preprocess",
    "postprocess",
    "processAllTokens"
  ])), _a);
  var Marked = class {
    constructor(...args) {
      __publicField(this, "defaults", _getDefaults());
      __publicField(this, "options", this.setOptions);
      __publicField(this, "parse", this.parseMarkdown(true));
      __publicField(this, "parseInline", this.parseMarkdown(false));
      __publicField(this, "Parser", _Parser);
      __publicField(this, "Renderer", _Renderer);
      __publicField(this, "TextRenderer", _TextRenderer);
      __publicField(this, "Lexer", _Lexer);
      __publicField(this, "Tokenizer", _Tokenizer);
      __publicField(this, "Hooks", _Hooks);
      this.use(...args);
    }
    /**
     * Run callback for every token
     */
    walkTokens(tokens, callback) {
      let values = [];
      for (const token of tokens) {
        values = values.concat(callback.call(this, token));
        switch (token.type) {
          case "table": {
            const tableToken = token;
            for (const cell of tableToken.header) {
              values = values.concat(this.walkTokens(cell.tokens, callback));
            }
            for (const row of tableToken.rows) {
              for (const cell of row) {
                values = values.concat(this.walkTokens(cell.tokens, callback));
              }
            }
            break;
          }
          case "list": {
            const listToken = token;
            values = values.concat(this.walkTokens(listToken.items, callback));
            break;
          }
          default: {
            const genericToken = token;
            if (this.defaults.extensions?.childTokens?.[genericToken.type]) {
              this.defaults.extensions.childTokens[genericToken.type].forEach((childTokens) => {
                const tokens2 = genericToken[childTokens].flat(Infinity);
                values = values.concat(this.walkTokens(tokens2, callback));
              });
            } else if (genericToken.tokens) {
              values = values.concat(this.walkTokens(genericToken.tokens, callback));
            }
          }
        }
      }
      return values;
    }
    use(...args) {
      const extensions = this.defaults.extensions || { renderers: {}, childTokens: {} };
      args.forEach((pack) => {
        const opts = { ...pack };
        opts.async = this.defaults.async || opts.async || false;
        if (pack.extensions) {
          pack.extensions.forEach((ext) => {
            if (!ext.name) {
              throw new Error("extension name required");
            }
            if ("renderer" in ext) {
              const prevRenderer = extensions.renderers[ext.name];
              if (prevRenderer) {
                extensions.renderers[ext.name] = function(...args2) {
                  let ret = ext.renderer.apply(this, args2);
                  if (ret === false) {
                    ret = prevRenderer.apply(this, args2);
                  }
                  return ret;
                };
              } else {
                extensions.renderers[ext.name] = ext.renderer;
              }
            }
            if ("tokenizer" in ext) {
              if (!ext.level || ext.level !== "block" && ext.level !== "inline") {
                throw new Error("extension level must be 'block' or 'inline'");
              }
              const extLevel = extensions[ext.level];
              if (extLevel) {
                extLevel.unshift(ext.tokenizer);
              } else {
                extensions[ext.level] = [ext.tokenizer];
              }
              if (ext.start) {
                if (ext.level === "block") {
                  if (extensions.startBlock) {
                    extensions.startBlock.push(ext.start);
                  } else {
                    extensions.startBlock = [ext.start];
                  }
                } else if (ext.level === "inline") {
                  if (extensions.startInline) {
                    extensions.startInline.push(ext.start);
                  } else {
                    extensions.startInline = [ext.start];
                  }
                }
              }
            }
            if ("childTokens" in ext && ext.childTokens) {
              extensions.childTokens[ext.name] = ext.childTokens;
            }
          });
          opts.extensions = extensions;
        }
        if (pack.renderer) {
          const renderer = this.defaults.renderer || new _Renderer(this.defaults);
          for (const prop in pack.renderer) {
            if (!(prop in renderer)) {
              throw new Error(`renderer '${prop}' does not exist`);
            }
            if (["options", "parser"].includes(prop)) {
              continue;
            }
            const rendererProp = prop;
            const rendererFunc = pack.renderer[rendererProp];
            const prevRenderer = renderer[rendererProp];
            renderer[rendererProp] = (...args2) => {
              let ret = rendererFunc.apply(renderer, args2);
              if (ret === false) {
                ret = prevRenderer.apply(renderer, args2);
              }
              return ret || "";
            };
          }
          opts.renderer = renderer;
        }
        if (pack.tokenizer) {
          const tokenizer = this.defaults.tokenizer || new _Tokenizer(this.defaults);
          for (const prop in pack.tokenizer) {
            if (!(prop in tokenizer)) {
              throw new Error(`tokenizer '${prop}' does not exist`);
            }
            if (["options", "rules", "lexer"].includes(prop)) {
              continue;
            }
            const tokenizerProp = prop;
            const tokenizerFunc = pack.tokenizer[tokenizerProp];
            const prevTokenizer = tokenizer[tokenizerProp];
            tokenizer[tokenizerProp] = (...args2) => {
              let ret = tokenizerFunc.apply(tokenizer, args2);
              if (ret === false) {
                ret = prevTokenizer.apply(tokenizer, args2);
              }
              return ret;
            };
          }
          opts.tokenizer = tokenizer;
        }
        if (pack.hooks) {
          const hooks = this.defaults.hooks || new _Hooks();
          for (const prop in pack.hooks) {
            if (!(prop in hooks)) {
              throw new Error(`hook '${prop}' does not exist`);
            }
            if (["options", "block"].includes(prop)) {
              continue;
            }
            const hooksProp = prop;
            const hooksFunc = pack.hooks[hooksProp];
            const prevHook = hooks[hooksProp];
            if (_Hooks.passThroughHooks.has(prop)) {
              hooks[hooksProp] = (arg) => {
                if (this.defaults.async) {
                  return Promise.resolve(hooksFunc.call(hooks, arg)).then((ret2) => {
                    return prevHook.call(hooks, ret2);
                  });
                }
                const ret = hooksFunc.call(hooks, arg);
                return prevHook.call(hooks, ret);
              };
            } else {
              hooks[hooksProp] = (...args2) => {
                let ret = hooksFunc.apply(hooks, args2);
                if (ret === false) {
                  ret = prevHook.apply(hooks, args2);
                }
                return ret;
              };
            }
          }
          opts.hooks = hooks;
        }
        if (pack.walkTokens) {
          const walkTokens2 = this.defaults.walkTokens;
          const packWalktokens = pack.walkTokens;
          opts.walkTokens = function(token) {
            let values = [];
            values.push(packWalktokens.call(this, token));
            if (walkTokens2) {
              values = values.concat(walkTokens2.call(this, token));
            }
            return values;
          };
        }
        this.defaults = { ...this.defaults, ...opts };
      });
      return this;
    }
    setOptions(opt) {
      this.defaults = { ...this.defaults, ...opt };
      return this;
    }
    lexer(src, options2) {
      return _Lexer.lex(src, options2 ?? this.defaults);
    }
    parser(tokens, options2) {
      return _Parser.parse(tokens, options2 ?? this.defaults);
    }
    parseMarkdown(blockType) {
      const parse2 = (src, options2) => {
        const origOpt = { ...options2 };
        const opt = { ...this.defaults, ...origOpt };
        const throwError = this.onError(!!opt.silent, !!opt.async);
        if (this.defaults.async === true && origOpt.async === false) {
          return throwError(new Error("marked(): The async option was set to true by an extension. Remove async: false from the parse options object to return a Promise."));
        }
        if (typeof src === "undefined" || src === null) {
          return throwError(new Error("marked(): input parameter is undefined or null"));
        }
        if (typeof src !== "string") {
          return throwError(new Error("marked(): input parameter is of type " + Object.prototype.toString.call(src) + ", string expected"));
        }
        if (opt.hooks) {
          opt.hooks.options = opt;
          opt.hooks.block = blockType;
        }
        const lexer2 = opt.hooks ? opt.hooks.provideLexer() : blockType ? _Lexer.lex : _Lexer.lexInline;
        const parser2 = opt.hooks ? opt.hooks.provideParser() : blockType ? _Parser.parse : _Parser.parseInline;
        if (opt.async) {
          return Promise.resolve(opt.hooks ? opt.hooks.preprocess(src) : src).then((src2) => lexer2(src2, opt)).then((tokens) => opt.hooks ? opt.hooks.processAllTokens(tokens) : tokens).then((tokens) => opt.walkTokens ? Promise.all(this.walkTokens(tokens, opt.walkTokens)).then(() => tokens) : tokens).then((tokens) => parser2(tokens, opt)).then((html22) => opt.hooks ? opt.hooks.postprocess(html22) : html22).catch(throwError);
        }
        try {
          if (opt.hooks) {
            src = opt.hooks.preprocess(src);
          }
          let tokens = lexer2(src, opt);
          if (opt.hooks) {
            tokens = opt.hooks.processAllTokens(tokens);
          }
          if (opt.walkTokens) {
            this.walkTokens(tokens, opt.walkTokens);
          }
          let html22 = parser2(tokens, opt);
          if (opt.hooks) {
            html22 = opt.hooks.postprocess(html22);
          }
          return html22;
        } catch (e) {
          return throwError(e);
        }
      };
      return parse2;
    }
    onError(silent, async) {
      return (e) => {
        e.message += "\nPlease report this to https://github.com/markedjs/marked.";
        if (silent) {
          const msg = "<p>An error occurred:</p><pre>" + escape2(e.message + "", true) + "</pre>";
          if (async) {
            return Promise.resolve(msg);
          }
          return msg;
        }
        if (async) {
          return Promise.reject(e);
        }
        throw e;
      };
    }
  };
  var markedInstance = new Marked();
  function marked(src, opt) {
    return markedInstance.parse(src, opt);
  }
  marked.options = marked.setOptions = function(options2) {
    markedInstance.setOptions(options2);
    marked.defaults = markedInstance.defaults;
    changeDefaults(marked.defaults);
    return marked;
  };
  marked.getDefaults = _getDefaults;
  marked.defaults = _defaults;
  marked.use = function(...args) {
    markedInstance.use(...args);
    marked.defaults = markedInstance.defaults;
    changeDefaults(marked.defaults);
    return marked;
  };
  marked.walkTokens = function(tokens, callback) {
    return markedInstance.walkTokens(tokens, callback);
  };
  marked.parseInline = markedInstance.parseInline;
  marked.Parser = _Parser;
  marked.parser = _Parser.parse;
  marked.Renderer = _Renderer;
  marked.TextRenderer = _TextRenderer;
  marked.Lexer = _Lexer;
  marked.lexer = _Lexer.lex;
  marked.Tokenizer = _Tokenizer;
  marked.Hooks = _Hooks;
  marked.parse = marked;
  var options = marked.options;
  var setOptions = marked.setOptions;
  var use = marked.use;
  var walkTokens = marked.walkTokens;
  var parseInline = marked.parseInline;
  var parser = _Parser.parse;
  var lexer = _Lexer.lex;

  // ../../../web/node_modules/dompurify/dist/purify.es.mjs
  function _OverloadYield(e, d) {
    this.v = e, this.k = d;
  }
  function _arrayLikeToArray(r, a) {
    (null == a || a > r.length) && (a = r.length);
    for (var e = 0, n = Array(a); e < a; e++) n[e] = r[e];
    return n;
  }
  function _arrayWithHoles(r) {
    if (Array.isArray(r)) return r;
  }
  function _iterableToArrayLimit(r, l) {
    var t2 = null == r ? null : "undefined" != typeof Symbol && r[Symbol.iterator] || r["@@iterator"];
    if (null != t2) {
      var e, n, i, u, a = [], f = true, o = false;
      try {
        if (i = (t2 = t2.call(r)).next, 0 === l) {
          if (Object(t2) !== t2) return;
          f = false;
        } else for (; !(f = (e = i.call(t2)).done) && (a.push(e.value), a.length !== l); f = true) ;
      } catch (r2) {
        o = true, n = r2;
      } finally {
        try {
          if (!f && null != t2.return && (u = t2.return(), Object(u) !== u)) return;
        } finally {
          if (o) throw n;
        }
      }
      return a;
    }
  }
  function _nonIterableRest() {
    throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
  }
  function _slicedToArray(r, e) {
    return _arrayWithHoles(r) || _iterableToArrayLimit(r, e) || _unsupportedIterableToArray(r, e) || _nonIterableRest();
  }
  function _unsupportedIterableToArray(r, a) {
    if (r) {
      if ("string" == typeof r) return _arrayLikeToArray(r, a);
      var t2 = {}.toString.call(r).slice(8, -1);
      return "Object" === t2 && r.constructor && (t2 = r.constructor.name), "Map" === t2 || "Set" === t2 ? Array.from(r) : "Arguments" === t2 || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(t2) ? _arrayLikeToArray(r, a) : void 0;
    }
  }
  function AsyncGenerator(e) {
    var t2, n;
    function resume(t3, n2) {
      try {
        var r = e[t3](n2), o = r.value, u = o instanceof _OverloadYield;
        Promise.resolve(u ? o.v : o).then(function(n3) {
          if (u) {
            var i = "return" === t3 && o.k ? t3 : "next";
            if (!o.k || n3.done) return resume(i, n3);
            n3 = e[i](n3).value;
          }
          settle(!!r.done, n3);
        }, function(e2) {
          resume("throw", e2);
        });
      } catch (e2) {
        settle(2, e2);
      }
    }
    function settle(e2, r) {
      2 === e2 ? t2.reject(r) : t2.resolve({
        value: r,
        done: e2
      }), (t2 = t2.next) ? resume(t2.key, t2.arg) : n = null;
    }
    this._invoke = function(e2, r) {
      return new Promise(function(o, u) {
        var i = {
          key: e2,
          arg: r,
          resolve: o,
          reject: u,
          next: null
        };
        n ? n = n.next = i : (t2 = n = i, resume(e2, r));
      });
    }, "function" != typeof e.return && (this.return = void 0);
  }
  AsyncGenerator.prototype["function" == typeof Symbol && Symbol.asyncIterator || "@@asyncIterator"] = function() {
    return this;
  }, AsyncGenerator.prototype.next = function(e) {
    return this._invoke("next", e);
  }, AsyncGenerator.prototype.throw = function(e) {
    return this._invoke("throw", e);
  }, AsyncGenerator.prototype.return = function(e) {
    return this._invoke("return", e);
  };
  var entries = Object.entries;
  var objectKeys = Object.keys;
  var setPrototypeOf = Object.setPrototypeOf;
  var isFrozen = Object.isFrozen;
  var getPrototypeOf = Object.getPrototypeOf;
  var getOwnPropertyDescriptor = Object.getOwnPropertyDescriptor;
  var freeze = Object.freeze;
  var seal = Object.seal;
  var create = Object.create;
  var _ref = typeof Reflect !== "undefined" && Reflect;
  var apply = _ref.apply;
  var construct = _ref.construct;
  if (!freeze) freeze = function freeze2(x) {
    return x;
  };
  if (!seal) seal = function seal2(x) {
    return x;
  };
  if (!apply) apply = function apply2(func, thisArg) {
    for (var _len = arguments.length, args = new Array(_len > 2 ? _len - 2 : 0), _key = 2; _key < _len; _key++) args[_key - 2] = arguments[_key];
    return func.apply(thisArg, args);
  };
  if (!construct) construct = function construct2(Func) {
    for (var _len2 = arguments.length, args = new Array(_len2 > 1 ? _len2 - 1 : 0), _key2 = 1; _key2 < _len2; _key2++) args[_key2 - 1] = arguments[_key2];
    return new Func(...args);
  };
  var arrayForEach = unapply(Array.prototype.forEach);
  var arrayLastIndexOf = unapply(Array.prototype.lastIndexOf);
  var arrayPop = unapply(Array.prototype.pop);
  var arrayPush = unapply(Array.prototype.push);
  var arraySlice = unapply(Array.prototype.slice);
  var arraySplice = unapply(Array.prototype.splice);
  var arrayIsArray = Array.isArray;
  var setCreate = unconstruct(Set);
  var setAdd = unapply(Set.prototype.add);
  var setHas = unapply(Set.prototype.has);
  var setDelete = unapply(Set.prototype.delete);
  var mapCreate = unconstruct(Map);
  var mapGet = unapply(Map.prototype.get);
  var mapSet = unapply(Map.prototype.set);
  var stringToLowerCase = unapply(String.prototype.toLowerCase);
  var stringToString = unapply(String.prototype.toString);
  var stringMatch = unapply(String.prototype.match);
  var stringReplace = unapply(String.prototype.replace);
  var stringIndexOf = unapply(String.prototype.indexOf);
  var stringTrim = unapply(String.prototype.trim);
  var numberToString = unapply(Number.prototype.toString);
  var booleanToString = unapply(Boolean.prototype.toString);
  var bigintToString = typeof BigInt === "undefined" ? null : unapply(BigInt.prototype.toString);
  var symbolToString = typeof Symbol === "undefined" ? null : unapply(Symbol.prototype.toString);
  var objectHasOwnProperty = unapply(Object.prototype.hasOwnProperty);
  var objectToString = unapply(Object.prototype.toString);
  var regExpTestRaw = RegExp.prototype.test;
  var regExpTest = function regExpTest2(regExp, value) {
    if (regExp instanceof RegExp) regExp.lastIndex = 0;
    return apply(regExpTestRaw, regExp, [value]);
  };
  var typeErrorCreate = unconstruct(TypeError);
  function unapply(func) {
    return function(thisArg) {
      for (var _len3 = arguments.length, args = new Array(_len3 > 1 ? _len3 - 1 : 0), _key3 = 1; _key3 < _len3; _key3++) args[_key3 - 1] = arguments[_key3];
      return apply(func, thisArg, args);
    };
  }
  function unconstruct(Func) {
    return function() {
      for (var _len4 = arguments.length, args = new Array(_len4), _key4 = 0; _key4 < _len4; _key4++) args[_key4] = arguments[_key4];
      return construct(Func, args);
    };
  }
  function addToSet(set2, array) {
    let transformCaseFunc = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : stringToLowerCase;
    if (setPrototypeOf) setPrototypeOf(set2, null);
    if (!arrayIsArray(array)) return set2;
    let l = array.length;
    while (l--) {
      let element = array[l];
      if (typeof element === "string") {
        const lcElement = transformCaseFunc(element);
        if (lcElement !== element) {
          if (!isFrozen(array)) array[l] = lcElement;
          element = lcElement;
        }
      }
      set2[element] = true;
    }
    return set2;
  }
  function cleanArray(array) {
    for (let index = 0; index < array.length; index++) if (!objectHasOwnProperty(array, index)) array[index] = null;
    return array;
  }
  function clone(object) {
    const newObject = create(null);
    for (const _ref2 of entries(object)) {
      var _ref3 = _slicedToArray(_ref2, 2);
      const property = _ref3[0];
      const value = _ref3[1];
      if (objectHasOwnProperty(object, property)) {
        if (arrayIsArray(value)) newObject[property] = cleanArray(value);
        else if (value && typeof value === "object" && value.constructor === Object) newObject[property] = clone(value);
        else newObject[property] = value;
      }
    }
    return newObject;
  }
  function cloneSet(set2) {
    const newSet = create(null);
    const keys = objectKeys(set2);
    let l = keys.length;
    while (l--) newSet[keys[l]] = set2[keys[l]];
    return newSet;
  }
  function stringifyValue(value) {
    switch (typeof value) {
      case "string":
        return value;
      case "number":
        return numberToString(value);
      case "boolean":
        return booleanToString(value);
      case "bigint":
        return bigintToString ? bigintToString(value) : "0";
      case "symbol":
        return symbolToString ? symbolToString(value) : "Symbol()";
      case "undefined":
        return objectToString(value);
      case "function":
      case "object": {
        if (value === null) return objectToString(value);
        const valueAsRecord = value;
        const valueToString = lookupGetter(valueAsRecord, "toString");
        if (typeof valueToString === "function") {
          const stringified = valueToString(valueAsRecord);
          return typeof stringified === "string" ? stringified : objectToString(stringified);
        }
        return objectToString(value);
      }
      default:
        return objectToString(value);
    }
  }
  function lookupGetter(object, prop) {
    while (object !== null) {
      const desc = getOwnPropertyDescriptor(object, prop);
      if (desc) {
        if (desc.get) return unapply(desc.get);
        if (typeof desc.value === "function") return unapply(desc.value);
      }
      object = getPrototypeOf(object);
    }
    function fallbackValue() {
      return null;
    }
    return fallbackValue;
  }
  function lookupAccessor(object, prop) {
    while (object !== null) {
      const desc = getOwnPropertyDescriptor(object, prop);
      if (desc) return desc.get ? unapply(desc.get) : null;
      object = getPrototypeOf(object);
    }
    return null;
  }
  function isRegex(value) {
    if (!value) return false;
    try {
      regExpTest(value, "");
      return true;
    } catch (_unused) {
      return false;
    }
  }
  var html$1 = freeze([
    "a",
    "abbr",
    "acronym",
    "address",
    "area",
    "article",
    "aside",
    "audio",
    "b",
    "bdi",
    "bdo",
    "big",
    "blink",
    "blockquote",
    "body",
    "br",
    "button",
    "canvas",
    "caption",
    "center",
    "cite",
    "code",
    "col",
    "colgroup",
    "content",
    "data",
    "datalist",
    "dd",
    "decorator",
    "del",
    "details",
    "dfn",
    "dialog",
    "dir",
    "div",
    "dl",
    "dt",
    "element",
    "em",
    "fieldset",
    "figcaption",
    "figure",
    "font",
    "footer",
    "form",
    "h1",
    "h2",
    "h3",
    "h4",
    "h5",
    "h6",
    "head",
    "header",
    "hgroup",
    "hr",
    "html",
    "i",
    "img",
    "input",
    "ins",
    "kbd",
    "label",
    "legend",
    "li",
    "main",
    "map",
    "mark",
    "marquee",
    "menu",
    "menuitem",
    "meter",
    "nav",
    "nobr",
    "ol",
    "optgroup",
    "option",
    "output",
    "p",
    "picture",
    "pre",
    "progress",
    "q",
    "rp",
    "rt",
    "ruby",
    "s",
    "samp",
    "search",
    "section",
    "select",
    "shadow",
    "slot",
    "small",
    "source",
    "spacer",
    "span",
    "strike",
    "strong",
    "style",
    "sub",
    "summary",
    "sup",
    "table",
    "tbody",
    "td",
    "template",
    "textarea",
    "tfoot",
    "th",
    "thead",
    "time",
    "tr",
    "track",
    "tt",
    "u",
    "ul",
    "var",
    "video",
    "wbr"
  ]);
  var svg$1 = freeze([
    "svg",
    "a",
    "altglyph",
    "altglyphdef",
    "altglyphitem",
    "animatecolor",
    "animatemotion",
    "animatetransform",
    "circle",
    "clippath",
    "defs",
    "desc",
    "ellipse",
    "filter",
    "font",
    "g",
    "glyph",
    "glyphref",
    "hkern",
    "image",
    "line",
    "lineargradient",
    "marker",
    "mask",
    "metadata",
    "mpath",
    "path",
    "pattern",
    "polygon",
    "polyline",
    "radialgradient",
    "rect",
    "stop",
    "style",
    "switch",
    "symbol",
    "text",
    "textpath",
    "title",
    "tref",
    "tspan",
    "view",
    "vkern"
  ]);
  var svgFilters = freeze([
    "feBlend",
    "feColorMatrix",
    "feComponentTransfer",
    "feComposite",
    "feConvolveMatrix",
    "feDiffuseLighting",
    "feDisplacementMap",
    "feDistantLight",
    "feDropShadow",
    "feFlood",
    "feFuncA",
    "feFuncB",
    "feFuncG",
    "feFuncR",
    "feGaussianBlur",
    "feImage",
    "feMerge",
    "feMergeNode",
    "feMorphology",
    "feOffset",
    "fePointLight",
    "feSpecularLighting",
    "feSpotLight",
    "feTile",
    "feTurbulence"
  ]);
  var svgDisallowed = freeze([
    "animate",
    "color-profile",
    "cursor",
    "discard",
    "font-face",
    "font-face-format",
    "font-face-name",
    "font-face-src",
    "font-face-uri",
    "foreignobject",
    "hatch",
    "hatchpath",
    "mesh",
    "meshgradient",
    "meshpatch",
    "meshrow",
    "missing-glyph",
    "script",
    "set",
    "solidcolor",
    "unknown",
    "use"
  ]);
  var mathMl$1 = freeze([
    "math",
    "menclose",
    "merror",
    "mfenced",
    "mfrac",
    "mglyph",
    "mi",
    "mlabeledtr",
    "mmultiscripts",
    "mn",
    "mo",
    "mover",
    "mpadded",
    "mphantom",
    "mroot",
    "mrow",
    "ms",
    "mspace",
    "msqrt",
    "mstyle",
    "msub",
    "msup",
    "msubsup",
    "mtable",
    "mtd",
    "mtext",
    "mtr",
    "munder",
    "munderover",
    "mprescripts"
  ]);
  var mathMlDisallowed = freeze([
    "maction",
    "maligngroup",
    "malignmark",
    "mlongdiv",
    "mscarries",
    "mscarry",
    "msgroup",
    "mstack",
    "msline",
    "msrow",
    "semantics",
    "annotation",
    "annotation-xml",
    "mprescripts",
    "none"
  ]);
  var text = freeze(["#text"]);
  var html2 = freeze([
    "accept",
    "action",
    "align",
    "alt",
    "autocapitalize",
    "autocomplete",
    "autopictureinpicture",
    "autoplay",
    "background",
    "bgcolor",
    "border",
    "capture",
    "cellpadding",
    "cellspacing",
    "checked",
    "cite",
    "class",
    "clear",
    "color",
    "cols",
    "colspan",
    "command",
    "commandfor",
    "controls",
    "controlslist",
    "coords",
    "crossorigin",
    "datetime",
    "decoding",
    "default",
    "dir",
    "disabled",
    "disablepictureinpicture",
    "disableremoteplayback",
    "download",
    "draggable",
    "enctype",
    "enterkeyhint",
    "exportparts",
    "face",
    "for",
    "headers",
    "height",
    "hidden",
    "high",
    "href",
    "hreflang",
    "id",
    "inert",
    "inputmode",
    "integrity",
    "ismap",
    "kind",
    "label",
    "lang",
    "list",
    "loading",
    "loop",
    "low",
    "max",
    "maxlength",
    "media",
    "method",
    "min",
    "minlength",
    "multiple",
    "muted",
    "name",
    "nonce",
    "noshade",
    "novalidate",
    "nowrap",
    "open",
    "optimum",
    "part",
    "pattern",
    "placeholder",
    "playsinline",
    "popover",
    "popovertarget",
    "popovertargetaction",
    "poster",
    "preload",
    "pubdate",
    "radiogroup",
    "readonly",
    "rel",
    "required",
    "rev",
    "reversed",
    "role",
    "rows",
    "rowspan",
    "spellcheck",
    "scope",
    "selected",
    "shape",
    "size",
    "sizes",
    "slot",
    "span",
    "srclang",
    "start",
    "src",
    "srcset",
    "step",
    "style",
    "summary",
    "tabindex",
    "title",
    "translate",
    "type",
    "usemap",
    "valign",
    "value",
    "width",
    "wrap",
    "xmlns"
  ]);
  var svg = freeze([
    "accent-height",
    "accumulate",
    "additive",
    "alignment-baseline",
    "amplitude",
    "ascent",
    "attributename",
    "attributetype",
    "azimuth",
    "basefrequency",
    "baseline-shift",
    "begin",
    "bias",
    "by",
    "class",
    "clip",
    "clippathunits",
    "clip-path",
    "clip-rule",
    "color",
    "color-interpolation",
    "color-interpolation-filters",
    "color-profile",
    "color-rendering",
    "cx",
    "cy",
    "d",
    "dx",
    "dy",
    "diffuseconstant",
    "direction",
    "display",
    "divisor",
    "dominant-baseline",
    "dur",
    "edgemode",
    "elevation",
    "end",
    "exponent",
    "fill",
    "fill-opacity",
    "fill-rule",
    "filter",
    "filterunits",
    "flood-color",
    "flood-opacity",
    "font-family",
    "font-size",
    "font-size-adjust",
    "font-stretch",
    "font-style",
    "font-variant",
    "font-weight",
    "fx",
    "fy",
    "g1",
    "g2",
    "glyph-name",
    "glyphref",
    "gradientunits",
    "gradienttransform",
    "height",
    "href",
    "id",
    "image-rendering",
    "in",
    "in2",
    "intercept",
    "k",
    "k1",
    "k2",
    "k3",
    "k4",
    "kerning",
    "keypoints",
    "keysplines",
    "keytimes",
    "lang",
    "lengthadjust",
    "letter-spacing",
    "kernelmatrix",
    "kernelunitlength",
    "lighting-color",
    "local",
    "marker-end",
    "marker-mid",
    "marker-start",
    "markerheight",
    "markerunits",
    "markerwidth",
    "maskcontentunits",
    "maskunits",
    "max",
    "mask",
    "mask-type",
    "media",
    "method",
    "mode",
    "min",
    "name",
    "numoctaves",
    "offset",
    "operator",
    "opacity",
    "order",
    "orient",
    "orientation",
    "origin",
    "overflow",
    "paint-order",
    "path",
    "pathlength",
    "patterncontentunits",
    "patterntransform",
    "patternunits",
    "pointer-events",
    "points",
    "preservealpha",
    "preserveaspectratio",
    "primitiveunits",
    "r",
    "rx",
    "ry",
    "radius",
    "refx",
    "refy",
    "repeatcount",
    "repeatdur",
    "restart",
    "result",
    "rotate",
    "scale",
    "seed",
    "shape-rendering",
    "slope",
    "specularconstant",
    "specularexponent",
    "spreadmethod",
    "startoffset",
    "stddeviation",
    "stitchtiles",
    "stop-color",
    "stop-opacity",
    "stroke-dasharray",
    "stroke-dashoffset",
    "stroke-linecap",
    "stroke-linejoin",
    "stroke-miterlimit",
    "stroke-opacity",
    "stroke",
    "stroke-width",
    "style",
    "surfacescale",
    "systemlanguage",
    "tabindex",
    "tablevalues",
    "targetx",
    "targety",
    "transform",
    "transform-origin",
    "text-anchor",
    "text-decoration",
    "text-orientation",
    "text-rendering",
    "textlength",
    "type",
    "u1",
    "u2",
    "unicode",
    "values",
    "vector-effect",
    "viewbox",
    "visibility",
    "version",
    "vert-adv-y",
    "vert-origin-x",
    "vert-origin-y",
    "width",
    "word-spacing",
    "wrap",
    "writing-mode",
    "xchannelselector",
    "ychannelselector",
    "x",
    "x1",
    "x2",
    "xmlns",
    "y",
    "y1",
    "y2",
    "z",
    "zoomandpan"
  ]);
  var mathMl = freeze([
    "accent",
    "accentunder",
    "align",
    "bevelled",
    "close",
    "columnalign",
    "columnlines",
    "columnspacing",
    "columnspan",
    "denomalign",
    "depth",
    "dir",
    "display",
    "displaystyle",
    "encoding",
    "fence",
    "frame",
    "height",
    "href",
    "id",
    "largeop",
    "length",
    "linethickness",
    "lquote",
    "lspace",
    "mathbackground",
    "mathcolor",
    "mathsize",
    "mathvariant",
    "maxsize",
    "minsize",
    "movablelimits",
    "notation",
    "numalign",
    "open",
    "rowalign",
    "rowlines",
    "rowspacing",
    "rowspan",
    "rspace",
    "rquote",
    "scriptlevel",
    "scriptminsize",
    "scriptsizemultiplier",
    "selection",
    "separator",
    "separators",
    "stretchy",
    "subscriptshift",
    "supscriptshift",
    "symmetric",
    "voffset",
    "width",
    "xmlns"
  ]);
  var xml = freeze([
    "xlink:href",
    "xml:id",
    "xlink:title",
    "xml:space",
    "xmlns:xlink"
  ]);
  var MUSTACHE_EXPR = seal(/{{[\w\W]*|^[\w\W]*}}/g);
  var ERB_EXPR = seal(/<%[\w\W]*|^[\w\W]*%>/g);
  var TMPLIT_EXPR = seal(/\${[\w\W]*/g);
  var TMPL_MARKER = seal(/{{|}}|\${|<%|%>/);
  var NON_WHITESPACE = seal(/[^\t\n\f\r ]/);
  var NUL = seal(/\u0000/g);
  var DATA_ATTR = seal(/^data-[\-\w.\u00B7-\uFFFF]+$/);
  var ARIA_ATTR = seal(/^aria-[\-\w]+$/);
  var IS_ALLOWED_URI = seal(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i);
  var IS_SCRIPT_OR_DATA = seal(/^(?:\w+script|data):/i);
  var ATTR_WHITESPACE = seal(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g);
  var DOCTYPE_NAME = seal(/^html$/i);
  var CUSTOM_ELEMENT = seal(/^[a-z][.\w]*(-[.\w]+)+$/i);
  var ELEMENT_MARKUP_PROBE = seal(/<[/\w!]/g);
  var COMMENT_MARKUP_PROBE = seal(/<[/\w]/g);
  var COMMENT_UNSAFE = seal(/<[/\w]|^-?>|--!?>|<!-$/g);
  var FALLBACK_TAG_CLOSE = seal(/<\/no(script|embed|frames)/i);
  var SELF_CLOSING_TAG = seal(/\/>/i);
  var ATTR_VALUE_BREAKOUT = seal(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i);
  var LEADING_WHITESPACE = seal(/^[\r\n\t ]+/);
  var NODE_TYPE = {
    element: 1,
    text: 3,
    processingInstruction: 7,
    comment: 8,
    document: 9,
    documentFragment: 11
  };
  var LITERAL_TEXT_ELEMENTS = freeze(addToSet({}, [
    "style",
    "script",
    "xmp",
    "iframe",
    "noembed",
    "noframes",
    "plaintext",
    "noscript"
  ]));
  var ESCAPABLE_TEXT_ELEMENTS = freeze(addToSet({}, ["textarea", "title"]));
  var VOID_ELEMENTS = freeze(addToSet({}, [
    "area",
    "base",
    "basefont",
    "bgsound",
    "br",
    "col",
    "embed",
    "frame",
    "hr",
    "img",
    "input",
    "keygen",
    "link",
    "meta",
    "param",
    "source",
    "track",
    "wbr"
  ]));
  var getGlobal = function getGlobal2() {
    return typeof window === "undefined" ? null : window;
  };
  var _createTrustedTypesPolicy = function _createTrustedTypesPolicy2(trustedTypes, purifyHostElement) {
    if (typeof trustedTypes !== "object" || typeof trustedTypes.createPolicy !== "function") return null;
    let suffix = null;
    const ATTR_NAME = "data-tt-policy-suffix";
    if (purifyHostElement && purifyHostElement.hasAttribute(ATTR_NAME)) suffix = purifyHostElement.getAttribute(ATTR_NAME);
    const policyName = "dompurify" + (suffix ? "#" + suffix : "");
    try {
      return trustedTypes.createPolicy(policyName, {
        createHTML(html3) {
          return html3;
        },
        createScriptURL(scriptUrl) {
          return scriptUrl;
        }
      });
    } catch (_) {
      console.warn("TrustedTypes policy " + policyName + " could not be created.");
      return null;
    }
  };
  var _createHooksMap = function _createHooksMap2() {
    return {
      afterSanitizeAttributes: [],
      afterSanitizeElements: [],
      afterSanitizeShadowDOM: [],
      beforeSanitizeAttributes: [],
      beforeSanitizeElements: [],
      beforeSanitizeShadowDOM: [],
      uponSanitizeAttribute: [],
      uponSanitizeElement: [],
      uponSanitizeShadowNode: []
    };
  };
  var _resolveSetOption = function _resolveSetOption2(cfg, key, fallback, options2) {
    return objectHasOwnProperty(cfg, key) && arrayIsArray(cfg[key]) ? addToSet(options2.base ? cloneSet(options2.base) : {}, cfg[key], options2.transform) : fallback;
  };
  var _resolveObjectOption = function _resolveObjectOption2(cfg, key, fallback) {
    const value = objectHasOwnProperty(cfg, key) ? cfg[key] : void 0;
    return value && typeof value === "object" ? clone(value) : fallback;
  };
  function createDOMPurify() {
    let window2 = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : getGlobal();
    const DOMPurify = (root) => createDOMPurify(root);
    DOMPurify.version = "3.5.0";
    DOMPurify.removed = [];
    if (!window2 || !window2.document || window2.document.nodeType !== NODE_TYPE.document || !window2.Element) {
      DOMPurify.isSupported = false;
      return DOMPurify;
    }
    let document2 = window2.document;
    const originalDocument = document2;
    const currentScript = originalDocument.currentScript;
    const HTMLTemplateElement = window2.HTMLTemplateElement, Node = window2.Node, Element = window2.Element, NodeFilter = window2.NodeFilter, DOMParser = window2.DOMParser, trustedTypes = window2.trustedTypes;
    const ElementPrototype = Element.prototype;
    const cloneNode = lookupGetter(ElementPrototype, "cloneNode");
    const remove = lookupGetter(ElementPrototype, "remove");
    const removeChild = lookupGetter(ElementPrototype, "removeChild");
    const removeAttributeNode = lookupGetter(ElementPrototype, "removeAttributeNode");
    const getNextSibling = lookupGetter(ElementPrototype, "nextSibling");
    const getChildNodes = lookupGetter(ElementPrototype, "childNodes");
    const getParentNode = lookupGetter(ElementPrototype, "parentNode");
    const getShadowRoot = lookupGetter(ElementPrototype, "shadowRoot");
    const getAttributes = lookupGetter(ElementPrototype, "attributes");
    const getNodeType = Node && Node.prototype ? lookupGetter(Node.prototype, "nodeType") : null;
    const getNodeName = Node && Node.prototype ? lookupGetter(Node.prototype, "nodeName") : null;
    const getOwnerDocument = Node && Node.prototype ? lookupGetter(Node.prototype, "ownerDocument") : null;
    const _readNodeType = function _readNodeType2(node) {
      return getNodeType ? getNodeType(node) : node.nodeType;
    };
    const _readNodeName = function _readNodeName2(node) {
      return getNodeName ? getNodeName(node) : node.nodeName;
    };
    const getNamespaceURI = lookupAccessor(ElementPrototype, "namespaceURI");
    const _readNamespaceURI = function _readNamespaceURI2(element) {
      return getNamespaceURI ? getNamespaceURI(element) : element.namespaceURI;
    };
    const _readAttributes = function _readAttributes2(node) {
      return _readNodeType(node) === NODE_TYPE.element ? getAttributes(node) : null;
    };
    if (typeof HTMLTemplateElement === "function") {
      const template = document2.createElement("template");
      if (template.content && template.content.ownerDocument) document2 = template.content.ownerDocument;
    }
    let trustedTypesPolicy;
    let emptyHTML = "";
    let defaultTrustedTypesPolicy;
    let defaultTrustedTypesPolicyResolved = false;
    let IN_TRUSTED_TYPES_POLICY = 0;
    const _assertNotInTrustedTypesPolicy = function _assertNotInTrustedTypesPolicy2() {
      if (IN_TRUSTED_TYPES_POLICY > 0) throw typeErrorCreate('A configured TRUSTED_TYPES_POLICY callback (createHTML or createScriptURL) must not call DOMPurify.sanitize, as that causes infinite recursion. Do not pass a policy whose callbacks wrap DOMPurify as TRUSTED_TYPES_POLICY; see the "DOMPurify and Trusted Types" section of the README.');
    };
    const _createTrustedHTML = function _createTrustedHTML2(html3) {
      _assertNotInTrustedTypesPolicy();
      IN_TRUSTED_TYPES_POLICY++;
      try {
        return trustedTypesPolicy.createHTML(html3);
      } finally {
        IN_TRUSTED_TYPES_POLICY--;
      }
    };
    const _signOutput = function _signOutput2(html3) {
      return trustedTypesPolicy && RETURN_TRUSTED_TYPE ? _createTrustedHTML(html3) : html3;
    };
    const _createTrustedScriptURL = function _createTrustedScriptURL2(scriptUrl) {
      _assertNotInTrustedTypesPolicy();
      IN_TRUSTED_TYPES_POLICY++;
      try {
        return trustedTypesPolicy.createScriptURL(scriptUrl);
      } finally {
        IN_TRUSTED_TYPES_POLICY--;
      }
    };
    const _getDefaultTrustedTypesPolicy = function _getDefaultTrustedTypesPolicy2() {
      if (!defaultTrustedTypesPolicyResolved) {
        defaultTrustedTypesPolicy = _createTrustedTypesPolicy(trustedTypes, currentScript);
        defaultTrustedTypesPolicyResolved = true;
      }
      return defaultTrustedTypesPolicy;
    };
    const _document = document2, implementation = _document.implementation, createNodeIterator = _document.createNodeIterator, createDocumentFragment = _document.createDocumentFragment, getElementsByTagName = _document.getElementsByTagName;
    const importNode = originalDocument.importNode;
    let hooks = _createHooksMap();
    DOMPurify.isSupported = typeof entries === "function" && typeof getParentNode === "function" && implementation && implementation.createHTMLDocument !== void 0;
    const MUSTACHE_EXPR$1 = MUSTACHE_EXPR, ERB_EXPR$1 = ERB_EXPR, TMPLIT_EXPR$1 = TMPLIT_EXPR, DATA_ATTR$1 = DATA_ATTR, ARIA_ATTR$1 = ARIA_ATTR, IS_SCRIPT_OR_DATA$1 = IS_SCRIPT_OR_DATA, ATTR_WHITESPACE$1 = ATTR_WHITESPACE, CUSTOM_ELEMENT$1 = CUSTOM_ELEMENT;
    let IS_ALLOWED_URI$1 = IS_ALLOWED_URI;
    let ALLOWED_TAGS = null;
    const DEFAULT_ALLOWED_TAGS = addToSet({}, [
      ...html$1,
      ...svg$1,
      ...svgFilters,
      ...mathMl$1,
      ...text
    ]);
    let ALLOWED_ATTR = null;
    const DEFAULT_ALLOWED_ATTR = addToSet({}, [
      ...html2,
      ...svg,
      ...mathMl,
      ...xml
    ]);
    let CUSTOM_ELEMENT_HANDLING = null;
    let FORBID_TAGS = null;
    let FORBID_ATTR = null;
    const EXTRA_ELEMENT_HANDLING = Object.seal(create(null, {
      tagCheck: {
        writable: true,
        configurable: false,
        enumerable: true,
        value: null
      },
      attributeCheck: {
        writable: true,
        configurable: false,
        enumerable: true,
        value: null
      }
    }));
    let ALLOW_ARIA_ATTR = true;
    let ALLOW_DATA_ATTR = true;
    let ALLOW_UNKNOWN_PROTOCOLS = false;
    let ALLOW_SELF_CLOSE_IN_ATTR = true;
    let SAFE_FOR_TEMPLATES = false;
    let SAFE_FOR_XML = true;
    let WHOLE_DOCUMENT = false;
    let SET_CONFIG = false;
    let SET_CONFIG_ALLOWED_TAGS = null;
    let SET_CONFIG_ALLOWED_ATTR = null;
    let FORCE_BODY = false;
    let RETURN_DOM = false;
    let RETURN_DOM_FRAGMENT = false;
    let RETURN_TRUSTED_TYPE = false;
    let SANITIZE_DOM = true;
    let SANITIZE_NAMED_PROPS = false;
    const SANITIZE_NAMED_PROPS_PREFIX = "user-content-";
    let KEEP_CONTENT = true;
    let IN_PLACE = false;
    let IN_PLACE_VISITED = null;
    let IN_PLACE_RETURNS = null;
    let USE_PROFILES = false;
    let FORBID_CONTENTS = null;
    const DEFAULT_FORBID_CONTENTS = addToSet({}, [
      "annotation-xml",
      "audio",
      "colgroup",
      "desc",
      "foreignobject",
      "head",
      "iframe",
      "math",
      "mi",
      "mn",
      "mo",
      "ms",
      "mtext",
      "noembed",
      "noframes",
      "noscript",
      "plaintext",
      "script",
      "selectedcontent",
      "style",
      "svg",
      "template",
      "thead",
      "title",
      "video",
      "xmp"
    ]);
    let DATA_URI_TAGS = null;
    const DEFAULT_DATA_URI_TAGS = addToSet({}, [
      "audio",
      "video",
      "img",
      "source",
      "image",
      "track"
    ]);
    let URI_SAFE_ATTRIBUTES = null;
    const DEFAULT_URI_SAFE_ATTRIBUTES = addToSet({}, [
      "alt",
      "class",
      "for",
      "id",
      "label",
      "name",
      "pattern",
      "placeholder",
      "role",
      "summary",
      "title",
      "value",
      "style",
      "xmlns"
    ]);
    const MATHML_NAMESPACE = "http://www.w3.org/1998/Math/MathML";
    const SVG_NAMESPACE = "http://www.w3.org/2000/svg";
    const HTML_NAMESPACE = "http://www.w3.org/1999/xhtml";
    let NAMESPACE = HTML_NAMESPACE;
    let IS_EMPTY_INPUT = false;
    let ALLOWED_NAMESPACES = null;
    const DEFAULT_ALLOWED_NAMESPACES = addToSet({}, [
      MATHML_NAMESPACE,
      SVG_NAMESPACE,
      HTML_NAMESPACE
    ], stringToString);
    const DEFAULT_MATHML_TEXT_INTEGRATION_POINTS = freeze(addToSet({}, [
      "mi",
      "mo",
      "mn",
      "ms",
      "mtext"
    ]));
    let MATHML_TEXT_INTEGRATION_POINTS = null;
    const DEFAULT_HTML_INTEGRATION_POINTS = freeze(addToSet({}, ["annotation-xml"]));
    let HTML_INTEGRATION_POINTS = null;
    const DEFAULT_CUSTOM_ELEMENT_HANDLING = freeze(create(null));
    const COMMON_SVG_AND_HTML_ELEMENTS = addToSet({}, [
      "title",
      "style",
      "font",
      "a",
      "script"
    ]);
    let PARSER_MEDIA_TYPE = null;
    const SUPPORTED_PARSER_MEDIA_TYPES = ["application/xhtml+xml", "text/html"];
    const DEFAULT_PARSER_MEDIA_TYPE = "text/html";
    let transformCaseFunc = null;
    let CONFIG = null;
    let ACTIVE_CONFIG = null;
    const formElement = document2.createElement("form");
    const isRegexOrFunction = function isRegexOrFunction2(testValue) {
      return testValue instanceof RegExp || testValue instanceof Function;
    };
    const _parseConfig = function _parseConfig2() {
      let cfg = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
      if (!cfg || typeof cfg !== "object") cfg = {};
      cfg = clone(cfg);
      PARSER_MEDIA_TYPE = SUPPORTED_PARSER_MEDIA_TYPES.indexOf(cfg.PARSER_MEDIA_TYPE) === -1 ? DEFAULT_PARSER_MEDIA_TYPE : cfg.PARSER_MEDIA_TYPE;
      transformCaseFunc = PARSER_MEDIA_TYPE === "application/xhtml+xml" ? stringToString : stringToLowerCase;
      ALLOWED_TAGS = _resolveSetOption(cfg, "ALLOWED_TAGS", DEFAULT_ALLOWED_TAGS, { transform: transformCaseFunc });
      ALLOWED_ATTR = _resolveSetOption(cfg, "ALLOWED_ATTR", DEFAULT_ALLOWED_ATTR, { transform: transformCaseFunc });
      ALLOWED_NAMESPACES = _resolveSetOption(cfg, "ALLOWED_NAMESPACES", DEFAULT_ALLOWED_NAMESPACES, { transform: stringToString });
      URI_SAFE_ATTRIBUTES = _resolveSetOption(cfg, "ADD_URI_SAFE_ATTR", DEFAULT_URI_SAFE_ATTRIBUTES, {
        transform: transformCaseFunc,
        base: DEFAULT_URI_SAFE_ATTRIBUTES
      });
      DATA_URI_TAGS = _resolveSetOption(cfg, "ADD_DATA_URI_TAGS", DEFAULT_DATA_URI_TAGS, {
        transform: transformCaseFunc,
        base: DEFAULT_DATA_URI_TAGS
      });
      FORBID_CONTENTS = _resolveSetOption(cfg, "FORBID_CONTENTS", DEFAULT_FORBID_CONTENTS, { transform: transformCaseFunc });
      FORBID_TAGS = _resolveSetOption(cfg, "FORBID_TAGS", create(null), { transform: transformCaseFunc });
      FORBID_ATTR = _resolveSetOption(cfg, "FORBID_ATTR", create(null), { transform: transformCaseFunc });
      USE_PROFILES = objectHasOwnProperty(cfg, "USE_PROFILES") ? cfg.USE_PROFILES && typeof cfg.USE_PROFILES === "object" ? clone(cfg.USE_PROFILES) : cfg.USE_PROFILES : false;
      ALLOW_ARIA_ATTR = cfg.ALLOW_ARIA_ATTR !== false;
      ALLOW_DATA_ATTR = cfg.ALLOW_DATA_ATTR !== false;
      ALLOW_UNKNOWN_PROTOCOLS = cfg.ALLOW_UNKNOWN_PROTOCOLS || false;
      ALLOW_SELF_CLOSE_IN_ATTR = cfg.ALLOW_SELF_CLOSE_IN_ATTR !== false;
      SAFE_FOR_TEMPLATES = cfg.SAFE_FOR_TEMPLATES || false;
      SAFE_FOR_XML = cfg.SAFE_FOR_XML !== false;
      WHOLE_DOCUMENT = cfg.WHOLE_DOCUMENT || false;
      RETURN_DOM = cfg.RETURN_DOM || false;
      RETURN_DOM_FRAGMENT = cfg.RETURN_DOM_FRAGMENT || false;
      RETURN_TRUSTED_TYPE = cfg.RETURN_TRUSTED_TYPE || false;
      FORCE_BODY = cfg.FORCE_BODY || false;
      SANITIZE_DOM = cfg.SANITIZE_DOM !== false;
      SANITIZE_NAMED_PROPS = cfg.SANITIZE_NAMED_PROPS || false;
      KEEP_CONTENT = cfg.KEEP_CONTENT !== false;
      IN_PLACE = cfg.IN_PLACE || false;
      IS_ALLOWED_URI$1 = isRegex(cfg.ALLOWED_URI_REGEXP) ? cfg.ALLOWED_URI_REGEXP : IS_ALLOWED_URI;
      NAMESPACE = typeof cfg.NAMESPACE === "string" ? cfg.NAMESPACE : HTML_NAMESPACE;
      MATHML_TEXT_INTEGRATION_POINTS = _resolveObjectOption(cfg, "MATHML_TEXT_INTEGRATION_POINTS", DEFAULT_MATHML_TEXT_INTEGRATION_POINTS);
      HTML_INTEGRATION_POINTS = _resolveObjectOption(cfg, "HTML_INTEGRATION_POINTS", DEFAULT_HTML_INTEGRATION_POINTS);
      const customElementHandling = _resolveObjectOption(cfg, "CUSTOM_ELEMENT_HANDLING", DEFAULT_CUSTOM_ELEMENT_HANDLING);
      CUSTOM_ELEMENT_HANDLING = create(null);
      if (objectHasOwnProperty(customElementHandling, "tagNameCheck") && isRegexOrFunction(customElementHandling.tagNameCheck)) CUSTOM_ELEMENT_HANDLING.tagNameCheck = customElementHandling.tagNameCheck;
      if (objectHasOwnProperty(customElementHandling, "attributeNameCheck") && isRegexOrFunction(customElementHandling.attributeNameCheck)) CUSTOM_ELEMENT_HANDLING.attributeNameCheck = customElementHandling.attributeNameCheck;
      if (objectHasOwnProperty(customElementHandling, "allowCustomizedBuiltInElements") && typeof customElementHandling.allowCustomizedBuiltInElements === "boolean") CUSTOM_ELEMENT_HANDLING.allowCustomizedBuiltInElements = customElementHandling.allowCustomizedBuiltInElements;
      seal(CUSTOM_ELEMENT_HANDLING);
      if (SAFE_FOR_TEMPLATES) ALLOW_DATA_ATTR = false;
      if (RETURN_DOM_FRAGMENT) RETURN_DOM = true;
      if (USE_PROFILES) {
        if ((objectHasOwnProperty(cfg, "ALLOWED_TAGS") || objectHasOwnProperty(cfg, "ALLOWED_ATTR")) && typeof console !== "undefined" && typeof console.warn === "function") console.warn("DOMPurify: USE_PROFILES overrides ALLOWED_TAGS and ALLOWED_ATTR. Do not use these options together.");
        ALLOWED_TAGS = addToSet({}, text);
        ALLOWED_ATTR = create(null);
        if (USE_PROFILES.html === true) {
          addToSet(ALLOWED_TAGS, html$1);
          addToSet(ALLOWED_ATTR, html2);
        }
        if (USE_PROFILES.svg === true) {
          addToSet(ALLOWED_TAGS, svg$1);
          addToSet(ALLOWED_ATTR, svg);
          addToSet(ALLOWED_ATTR, xml);
        }
        if (USE_PROFILES.svgFilters === true) {
          addToSet(ALLOWED_TAGS, svgFilters);
          addToSet(ALLOWED_ATTR, svg);
          addToSet(ALLOWED_ATTR, xml);
        }
        if (USE_PROFILES.mathMl === true) {
          addToSet(ALLOWED_TAGS, mathMl$1);
          addToSet(ALLOWED_ATTR, mathMl);
          addToSet(ALLOWED_ATTR, xml);
        }
      }
      EXTRA_ELEMENT_HANDLING.tagCheck = null;
      EXTRA_ELEMENT_HANDLING.attributeCheck = null;
      if (objectHasOwnProperty(cfg, "ADD_TAGS")) {
        if (typeof cfg.ADD_TAGS === "function") EXTRA_ELEMENT_HANDLING.tagCheck = cfg.ADD_TAGS;
        else if (arrayIsArray(cfg.ADD_TAGS)) {
          if (ALLOWED_TAGS === DEFAULT_ALLOWED_TAGS) ALLOWED_TAGS = cloneSet(ALLOWED_TAGS);
          addToSet(ALLOWED_TAGS, cfg.ADD_TAGS, transformCaseFunc);
        }
      }
      if (objectHasOwnProperty(cfg, "ADD_ATTR")) {
        if (typeof cfg.ADD_ATTR === "function") EXTRA_ELEMENT_HANDLING.attributeCheck = cfg.ADD_ATTR;
        else if (arrayIsArray(cfg.ADD_ATTR)) {
          if (ALLOWED_ATTR === DEFAULT_ALLOWED_ATTR) ALLOWED_ATTR = cloneSet(ALLOWED_ATTR);
          addToSet(ALLOWED_ATTR, cfg.ADD_ATTR, transformCaseFunc);
        }
      }
      if (objectHasOwnProperty(cfg, "ADD_FORBID_CONTENTS") && arrayIsArray(cfg.ADD_FORBID_CONTENTS)) {
        if (FORBID_CONTENTS === DEFAULT_FORBID_CONTENTS) FORBID_CONTENTS = cloneSet(FORBID_CONTENTS);
        addToSet(FORBID_CONTENTS, cfg.ADD_FORBID_CONTENTS, transformCaseFunc);
      }
      if (KEEP_CONTENT) ALLOWED_TAGS["#text"] = true;
      if (WHOLE_DOCUMENT) addToSet(ALLOWED_TAGS, [
        "html",
        "head",
        "body"
      ]);
      if (ALLOWED_TAGS.table) {
        ALLOWED_TAGS.tbody = true;
        delete FORBID_TAGS.tbody;
      }
      if (cfg.TRUSTED_TYPES_POLICY) {
        if (typeof cfg.TRUSTED_TYPES_POLICY.createHTML !== "function") throw typeErrorCreate('TRUSTED_TYPES_POLICY configuration option must provide a "createHTML" hook.');
        if (typeof cfg.TRUSTED_TYPES_POLICY.createScriptURL !== "function") throw typeErrorCreate('TRUSTED_TYPES_POLICY configuration option must provide a "createScriptURL" hook.');
        const previousTrustedTypesPolicy = trustedTypesPolicy;
        trustedTypesPolicy = cfg.TRUSTED_TYPES_POLICY;
        try {
          emptyHTML = _createTrustedHTML("");
        } catch (error) {
          trustedTypesPolicy = previousTrustedTypesPolicy;
          throw error;
        }
      } else if (cfg.TRUSTED_TYPES_POLICY === null) {
        trustedTypesPolicy = void 0;
        emptyHTML = "";
      } else {
        if (trustedTypesPolicy === void 0) trustedTypesPolicy = _getDefaultTrustedTypesPolicy();
        if (trustedTypesPolicy && typeof emptyHTML === "string") emptyHTML = _createTrustedHTML("");
      }
      CONFIG = freeze(cfg);
    };
    const ALL_SVG_TAGS = addToSet({}, [
      ...svg$1,
      ...svgFilters,
      ...svgDisallowed
    ]);
    const ALL_MATHML_TAGS = addToSet({}, [...mathMl$1, ...mathMlDisallowed]);
    const _checkSvgNamespace = function _checkSvgNamespace2(tagName, parentNamespace, parentTagName) {
      if (parentNamespace === HTML_NAMESPACE) return tagName === "svg";
      if (parentNamespace === MATHML_NAMESPACE) return tagName === "svg" && (parentTagName === "annotation-xml" || MATHML_TEXT_INTEGRATION_POINTS[parentTagName]);
      return Boolean(ALL_SVG_TAGS[tagName]);
    };
    const _checkMathMlNamespace = function _checkMathMlNamespace2(tagName, parentNamespace, parentTagName) {
      if (parentNamespace === HTML_NAMESPACE) return tagName === "math";
      if (parentNamespace === SVG_NAMESPACE) return tagName === "math" && HTML_INTEGRATION_POINTS[parentTagName];
      return Boolean(ALL_MATHML_TAGS[tagName]);
    };
    const _checkHtmlNamespace = function _checkHtmlNamespace2(tagName, parentNamespace, parentTagName) {
      if (parentNamespace === SVG_NAMESPACE && !HTML_INTEGRATION_POINTS[parentTagName]) return false;
      if (parentNamespace === MATHML_NAMESPACE && !MATHML_TEXT_INTEGRATION_POINTS[parentTagName]) return false;
      return !ALL_MATHML_TAGS[tagName] && (COMMON_SVG_AND_HTML_ELEMENTS[tagName] || !ALL_SVG_TAGS[tagName]);
    };
    const _checkValidNamespace = function _checkValidNamespace2(element) {
      const elementNamespace = _readNamespaceURI(element);
      const tagName = stringToLowerCase(_readNodeName(element));
      const parentNode = getParentNode(element);
      let parentNamespace = NAMESPACE;
      let parentTagName = "template";
      if (parentNode && _readNodeType(parentNode) === NODE_TYPE.element) {
        parentNamespace = _readNamespaceURI(parentNode);
        parentTagName = stringToLowerCase(_readNodeName(parentNode));
      }
      if (!ALLOWED_NAMESPACES[elementNamespace]) return false;
      if (elementNamespace === SVG_NAMESPACE) return _checkSvgNamespace(tagName, parentNamespace, parentTagName);
      if (elementNamespace === MATHML_NAMESPACE) return _checkMathMlNamespace(tagName, parentNamespace, parentTagName);
      if (elementNamespace === HTML_NAMESPACE) return _checkHtmlNamespace(tagName, parentNamespace, parentTagName);
      return PARSER_MEDIA_TYPE === "application/xhtml+xml";
    };
    const _forceRemove = function _forceRemove2(node) {
      arrayPush(DOMPurify.removed, { element: node });
      _unmarkVisited(node);
      try {
        getParentNode(node).removeChild(node);
      } catch (_) {
        remove(node);
        if (!getParentNode(node)) throw typeErrorCreate("a node selected for removal could not be detached from its tree and cannot be safely returned; refusing to sanitize in place");
      }
    };
    const _stripAttributeNode = function _stripAttributeNode2(element, attribute, name) {
      try {
        removeAttributeNode(element, attribute);
      } catch (_) {
        try {
          element.removeAttribute(name);
        } catch (_2) {
        }
      }
    };
    const _neutralizeRoot = function _neutralizeRoot2(root) {
      _neutralizeSubtree(root);
      const childNodes = getChildNodes(root);
      if (childNodes) arrayForEach(arraySlice(childNodes), (child) => {
        try {
          removeChild(root, child);
        } catch (_) {
        }
      });
      if (_readNodeType(root) === NODE_TYPE.element) _stripAttributes(root, false);
    };
    const _removeAttribute = function _removeAttribute2(name, element, attr) {
      if (!attr) try {
        attr = element.getAttributeNode(name);
      } catch (_) {
        attr = null;
      }
      arrayPush(DOMPurify.removed, {
        attribute: attr || null,
        from: element
      });
      if (attr) _stripAttributeNode(element, attr, name);
      else try {
        element.removeAttribute(name);
      } catch (_) {
      }
      if (name === "is") {
        if (RETURN_DOM || RETURN_DOM_FRAGMENT) try {
          _forceRemove(element);
        } catch (_) {
        }
        else try {
          element.setAttribute(name, "");
        } catch (_) {
        }
      }
    };
    const _stripAttributes = function _stripAttributes2(element, keepAllowed) {
      const attributes = getAttributes(element);
      if (!attributes) return;
      for (let i = attributes.length - 1; i >= 0; --i) {
        const attribute = attributes[i];
        const name = attribute && attribute.name;
        if (typeof name === "string" && !(keepAllowed && ALLOWED_ATTR[transformCaseFunc(name)])) _stripAttributeNode(element, attribute, name);
      }
    };
    const _walkTree = function _walkTree2(root, visit) {
      let intoShadow = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : false;
      const stack = [root];
      while (stack.length > 0) {
        const node = stack.pop();
        if (visit(node) === false) continue;
        if (intoShadow && _readNodeType(node) === NODE_TYPE.element) {
          const shadowRoot = getShadowRoot(node);
          if (_isDocumentFragment(shadowRoot)) stack.push(shadowRoot);
        }
        const childNodes = getChildNodes(node);
        if (childNodes) for (let i = childNodes.length - 1; i >= 0; --i) stack.push(childNodes[i]);
      }
    };
    const _neutralizeSubtree = function _neutralizeSubtree2(root) {
      _walkTree(root, (node) => {
        if (_readNodeType(node) === NODE_TYPE.element) _stripAttributes(node, true);
      });
    };
    const _collectLiveNodes = function _collectLiveNodes2(root) {
      const nodes = [];
      _walkTree(root, (node) => {
        nodes.push(node);
      }, true);
      return nodes;
    };
    const _neutralizeEscapedElements = function _neutralizeEscapedElements2(root, before) {
      const reachable = setCreate();
      arrayForEach(_collectLiveNodes(root), (node) => {
        if (_readNodeType(node) === NODE_TYPE.element) setAdd(reachable, node);
      });
      arrayForEach(before, (node) => {
        if (_readNodeType(node) === NODE_TYPE.element && !setHas(reachable, node)) _stripAttributes(node, true);
      });
    };
    const _neutralizeDiscarded = function _neutralizeDiscarded2(root, before) {
      let rootRemoved = false;
      arrayForEach(DOMPurify.removed, (entry) => {
        if (entry.element) {
          rootRemoved = rootRemoved || entry.element === root;
          _neutralizeSubtree(entry.element);
        }
      });
      _neutralizeEscapedElements(root, before);
      return rootRemoved;
    };
    const _failClosed = function _failClosed2(root, before) {
      _neutralizeRoot(root);
      _neutralizeDiscarded(root, before);
    };
    const MAX_UNVISITED_ROUNDS = 8;
    const _sanitizeUnvisitedNodes = function _sanitizeUnvisitedNodes2(root, before) {
      const visited = IN_PLACE_VISITED;
      let original = null;
      const getOriginal = () => {
        if (!original) {
          original = setCreate();
          arrayForEach(before, (beforeNode) => {
            setAdd(original, beforeNode);
          });
        }
        return original;
      };
      for (let round = 0; ; round++) {
        const live = _collectLiveNodes(root);
        const containers = [root];
        let pending = false;
        arrayForEach(live, (node) => {
          if (_readNodeType(node) === NODE_TYPE.documentFragment) arrayPush(containers, node);
          else if (!pending && !setHas(visited, node)) pending = setHas(getOriginal(), node);
        });
        if (!pending) return;
        if (round >= MAX_UNVISITED_ROUNDS) throw typeErrorCreate("the tree kept changing during in-place sanitization; refusing to return it");
        arrayForEach(containers, (container, index) => {
          _sanitizeUnvisitedIn(container, index > 0, root, getOriginal());
        });
      }
    };
    const _sanitizeUnvisitedIn = function _sanitizeUnvisitedIn2(container, inShadow, root, original) {
      const iterator = _createNodeIterator(container);
      let node;
      while (node = iterator.nextNode()) {
        if (!setHas(original, node) || setHas(IN_PLACE_VISITED, node)) continue;
        if (inShadow) _executeHooks(hooks.uponSanitizeShadowNode, node, null);
        _sanitizeNode(node, root);
      }
    };
    const _markVisited = function _markVisited2(node) {
      if (!IN_PLACE_VISITED) return;
      setAdd(IN_PLACE_VISITED, node);
      const returns = mapGet(IN_PLACE_RETURNS, node);
      if (returns === void 0) return;
      if (returns >= MAX_UNVISITED_ROUNDS) throw typeErrorCreate("the tree kept changing during in-place sanitization; refusing to return it");
      mapSet(IN_PLACE_RETURNS, node, returns + 1);
    };
    const _unmarkVisited = function _unmarkVisited2(node) {
      if (!IN_PLACE_VISITED) return;
      setDelete(IN_PLACE_VISITED, node);
      if (mapGet(IN_PLACE_RETURNS, node) === void 0) mapSet(IN_PLACE_RETURNS, node, 0);
    };
    const _isPatchLinkageAttribute = function _isPatchLinkageAttribute2(lcName, lcTag) {
      if (!SAFE_FOR_XML) return false;
      if (lcName === "patchsrc") return true;
      return lcName === "for" && lcTag !== "label" && lcTag !== "output";
    };
    const _neutralizePatchLinkage = function _neutralizePatchLinkage2(root) {
      if (!SAFE_FOR_XML) return;
      _walkTree(root, (node) => {
        const nodeType = _readNodeType(node);
        if (nodeType === NODE_TYPE.processingInstruction || nodeType === NODE_TYPE.comment && regExpTest(COMMENT_MARKUP_PROBE, node.data)) {
          try {
            remove(node);
          } catch (_) {
          }
          return false;
        }
        if (nodeType === NODE_TYPE.element) {
          const element = node;
          const lcTag = transformCaseFunc(_readNodeName(node));
          try {
            if (element.hasAttribute && element.hasAttribute("patchsrc")) element.removeAttribute("patchsrc");
            if (element.hasAttribute && element.hasAttribute("for") && _isPatchLinkageAttribute("for", lcTag)) element.removeAttribute("for");
          } catch (_) {
          }
        }
        return true;
      });
    };
    const _initDocument = function _initDocument2(dirty) {
      let doc = null;
      let leadingWhitespace = null;
      if (FORCE_BODY) dirty = "<remove></remove>" + dirty;
      else {
        const matches = stringMatch(dirty, LEADING_WHITESPACE);
        leadingWhitespace = matches && matches[0];
      }
      if (PARSER_MEDIA_TYPE === "application/xhtml+xml" && NAMESPACE === HTML_NAMESPACE) dirty = '<html xmlns="http://www.w3.org/1999/xhtml"><head></head><body>' + dirty + "</body></html>";
      const dirtyPayload = trustedTypesPolicy ? _createTrustedHTML(dirty) : dirty;
      if (NAMESPACE === HTML_NAMESPACE) try {
        doc = new DOMParser().parseFromString(dirtyPayload, PARSER_MEDIA_TYPE);
      } catch (_) {
      }
      if (!doc || !doc.documentElement) {
        doc = implementation.createDocument(NAMESPACE, "template", null);
        try {
          doc.documentElement.innerHTML = IS_EMPTY_INPUT ? emptyHTML : dirtyPayload;
        } catch (_) {
        }
      }
      const body = doc.body || doc.documentElement;
      if (dirty && leadingWhitespace) body.insertBefore(document2.createTextNode(leadingWhitespace), body.childNodes[0] || null);
      if (NAMESPACE === HTML_NAMESPACE) return getElementsByTagName.call(doc, WHOLE_DOCUMENT ? "html" : "body")[0];
      return WHOLE_DOCUMENT ? doc.documentElement : body;
    };
    const _createNodeIterator = function _createNodeIterator2(root) {
      let characterDataOnly = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : false;
      const doc = getOwnerDocument ? getOwnerDocument(root) : root.ownerDocument;
      return createNodeIterator.call(doc || root, root, (characterDataOnly ? 0 : NodeFilter.SHOW_ELEMENT) | NodeFilter.SHOW_COMMENT | NodeFilter.SHOW_TEXT | NodeFilter.SHOW_PROCESSING_INSTRUCTION | NodeFilter.SHOW_CDATA_SECTION, null);
    };
    const _stripTemplateExpressions = function _stripTemplateExpressions2(value) {
      value = stringReplace(value, MUSTACHE_EXPR$1, " ");
      value = stringReplace(value, ERB_EXPR$1, " ");
      value = stringReplace(value, TMPLIT_EXPR$1, " ");
      return value;
    };
    const _scrubTemplateExpressions2 = function _scrubTemplateExpressions(node) {
      var _node$querySelectorAl;
      _scrubOwnCharacterData(node);
      const templates = (_node$querySelectorAl = node.querySelectorAll) === null || _node$querySelectorAl === void 0 ? void 0 : _node$querySelectorAl.call(node, "template");
      if (templates) arrayForEach(templates, (tmpl) => {
        if (_isDocumentFragment(tmpl.content)) _scrubTemplateExpressions2(tmpl.content);
      });
    };
    const TABLE_STRUCTURE = "table,tbody,thead,tfoot,tr,colgroup";
    const TABLE_CONTENT = addToSet({}, [
      "caption",
      "colgroup",
      "col",
      "thead",
      "tbody",
      "tfoot",
      "tr",
      "td",
      "th",
      "style",
      "script",
      "template"
    ]);
    const _fosterTableContent = function _fosterTableContent2(root) {
      var _root$querySelectorAl;
      if (PARSER_MEDIA_TYPE === "application/xhtml+xml") return;
      const structures = (_root$querySelectorAl = root.querySelectorAll) === null || _root$querySelectorAl === void 0 ? void 0 : _root$querySelectorAl.call(root, TABLE_STRUCTURE);
      if (!structures || structures.length === 0) return;
      arrayForEach(structures, (structure) => {
        arrayForEach(arraySlice(getChildNodes(structure)), (child) => {
          const nodeType = _readNodeType(child);
          if (!(nodeType === NODE_TYPE.text ? regExpTest(NON_WHITESPACE, child.data) : nodeType === NODE_TYPE.element && !TABLE_CONTENT[stringToLowerCase(_readNodeName(child))])) return;
          let table = structure;
          while (table && stringToLowerCase(_readNodeName(table)) !== "table") table = table === root ? null : getParentNode(table);
          const tableParent = table && table !== root ? getParentNode(table) : null;
          if (tableParent) try {
            tableParent.insertBefore(child, table);
          } catch (_) {
          }
        });
      });
    };
    const _scrubOwnCharacterData = function _scrubOwnCharacterData2(node) {
      _fosterTableContent(node);
      node.normalize();
      const walker = _createNodeIterator(node, true);
      let previousText = null;
      let currentNode = walker.nextNode();
      while (currentNode) {
        let data = currentNode.data;
        if (regExpTest(NUL, data)) {
          data = stringReplace(data, NUL, "");
          currentNode.data = data;
        }
        if (regExpTest(TMPL_MARKER, data)) {
          data = _stripTemplateExpressions(data);
          currentNode.data = data;
        }
        if (data && _readNodeType(currentNode) === NODE_TYPE.text) {
          if (previousText) {
            const tail = previousText.data;
            const boundary = tail[tail.length - 1] + data[0];
            if (regExpTest(TMPL_MARKER, boundary)) previousText.data = _stripTemplateExpressions(tail + data[0]);
          }
          previousText = currentNode;
        }
        currentNode = walker.nextNode();
      }
    };
    const _isClobbered = function _isClobbered2(element) {
      const realTagName = getNodeName ? getNodeName(element) : null;
      if (typeof realTagName !== "string") return false;
      if (transformCaseFunc(realTagName) !== "form") return false;
      return typeof element.nodeName !== "string" || typeof element.textContent !== "string" || typeof element.removeChild !== "function" || element.attributes !== getAttributes(element) || typeof element.removeAttribute !== "function" || typeof element.removeAttributeNode !== "function" || typeof element.getAttributeNode !== "function" || typeof element.setAttribute !== "function" || typeof element.namespaceURI !== "string" || typeof element.insertBefore !== "function" || typeof element.hasChildNodes !== "function" || element.nodeType !== getNodeType(element) || element.childNodes !== getChildNodes(element);
    };
    const _nodeTypeOf = function _nodeTypeOf2(value) {
      if (!getNodeType || typeof value !== "object" || value === null) return null;
      try {
        const nodeType = getNodeType(value);
        return typeof nodeType === "number" ? nodeType : null;
      } catch (_) {
        return null;
      }
    };
    const _isDocumentFragment = function _isDocumentFragment2(value) {
      return _nodeTypeOf(value) === NODE_TYPE.documentFragment;
    };
    const _isNode = function _isNode2(value) {
      return _nodeTypeOf(value) !== null;
    };
    function _executeHooks(hooks2, currentNode, data) {
      if (hooks2.length === 0) return;
      arrayForEach(hooks2, (hook) => {
        hook.call(DOMPurify, currentNode, data, CONFIG);
      });
    }
    const _isMalformedElement = function _isMalformedElement2(currentNode, tagName) {
      const isVoid = Boolean(VOID_ELEMENTS[tagName]);
      const literal = Boolean(LITERAL_TEXT_ELEMENTS[tagName]);
      if (!isVoid && !literal && !ESCAPABLE_TEXT_ELEMENTS[tagName]) return false;
      if (_readNamespaceURI(currentNode) !== HTML_NAMESPACE) return false;
      if (isVoid) return currentNode.hasChildNodes();
      if (_isNode(currentNode.firstElementChild)) return true;
      return literal && typeof currentNode.textContent === "string" && regExpTest(ELEMENT_MARKUP_PROBE, currentNode.textContent);
    };
    const _isUnsafeNode = function _isUnsafeNode2(currentNode, tagName, nodeType) {
      if (SAFE_FOR_XML && nodeType === NODE_TYPE.element && currentNode.hasChildNodes() && !_isNode(currentNode.firstElementChild) && regExpTest(ELEMENT_MARKUP_PROBE, currentNode.textContent) && regExpTest(ELEMENT_MARKUP_PROBE, currentNode.innerHTML)) return true;
      if (SAFE_FOR_XML && nodeType === NODE_TYPE.element && _isMalformedElement(currentNode, tagName)) return true;
      if (nodeType === NODE_TYPE.processingInstruction) return true;
      if (SAFE_FOR_XML && nodeType === NODE_TYPE.comment && regExpTest(COMMENT_UNSAFE, currentNode.data)) return true;
      return false;
    };
    const _matchesNameCheck = function _matchesNameCheck2(check, name) {
      if (check instanceof RegExp) return regExpTest(check, name);
      if (check instanceof Function) {
        for (var _len = arguments.length, args = new Array(_len > 2 ? _len - 2 : 0), _key = 2; _key < _len; _key++) args[_key - 2] = arguments[_key];
        return Boolean(check(name, ...args));
      }
      return false;
    };
    const _sanitizeDisallowedNode = function _sanitizeDisallowedNode2(currentNode, tagName, root) {
      if (!FORBID_TAGS[tagName] && _isBasicCustomElement(tagName) && _matchesNameCheck(CUSTOM_ELEMENT_HANDLING.tagNameCheck, tagName)) return false;
      if (KEEP_CONTENT && !FORBID_CONTENTS[tagName] && !FORBID_CONTENTS[stringToLowerCase(tagName)]) {
        const parentNode = getParentNode(currentNode);
        const childNodes = getChildNodes(currentNode);
        if (childNodes && parentNode) {
          const childCount = childNodes.length;
          for (let i = childCount - 1; i >= 0; --i) {
            const hoisted = currentNode === root ? cloneNode(childNodes[i], true) : childNodes[i];
            parentNode.insertBefore(hoisted, getNextSibling(currentNode));
          }
        }
      }
      _forceRemove(currentNode);
      return true;
    };
    const _forkSharedAllowlist = function _forkSharedAllowlist2(hookList, set2, defaultSet, setConfigSet) {
      if (hookList.length === 0) return set2;
      return set2 === defaultSet || set2 === setConfigSet ? cloneSet(set2) : set2;
    };
    const _handleHookDetachedNode = function _handleHookDetachedNode2(currentNode, root) {
      if (currentNode === root || getParentNode(currentNode) !== null) return false;
      if (IN_PLACE) _neutralizeSubtree(currentNode);
      _unmarkVisited(currentNode);
      return true;
    };
    const _sanitizeElements = function _sanitizeElements2(currentNode, root) {
      _markVisited(currentNode);
      _executeHooks(hooks.beforeSanitizeElements, currentNode, null);
      if (_handleHookDetachedNode(currentNode, root)) return true;
      if (_isClobbered(currentNode)) {
        _forceRemove(currentNode);
        return true;
      }
      const tagName = transformCaseFunc(_readNodeName(currentNode));
      const nodeType = _readNodeType(currentNode);
      ALLOWED_TAGS = _forkSharedAllowlist(hooks.uponSanitizeElement, ALLOWED_TAGS, DEFAULT_ALLOWED_TAGS, SET_CONFIG_ALLOWED_TAGS);
      _executeHooks(hooks.uponSanitizeElement, currentNode, {
        tagName,
        allowedTags: ALLOWED_TAGS
      });
      if (_handleHookDetachedNode(currentNode, root)) return true;
      if (_isUnsafeNode(currentNode, tagName, nodeType)) {
        _forceRemove(currentNode);
        return true;
      }
      if ((FORBID_TAGS[tagName] || !(EXTRA_ELEMENT_HANDLING.tagCheck instanceof Function && EXTRA_ELEMENT_HANDLING.tagCheck(tagName)) && !ALLOWED_TAGS[tagName]) && _sanitizeDisallowedNode(currentNode, tagName, root)) return true;
      if (nodeType === NODE_TYPE.element && !_checkValidNamespace(currentNode)) {
        _forceRemove(currentNode);
        return true;
      }
      if ((tagName === "noscript" || tagName === "noembed" || tagName === "noframes") && regExpTest(FALLBACK_TAG_CLOSE, currentNode.innerHTML)) {
        _forceRemove(currentNode);
        return true;
      }
      if (SAFE_FOR_TEMPLATES && nodeType === NODE_TYPE.text) {
        const text2 = currentNode.textContent;
        const content = _stripTemplateExpressions(text2);
        if (text2 !== content) {
          arrayPush(DOMPurify.removed, { element: currentNode.cloneNode() });
          currentNode.textContent = content;
        }
      }
      _executeHooks(hooks.afterSanitizeElements, currentNode, null);
      return _handleHookDetachedNode(currentNode, root);
    };
    const _isValidAttribute = function _isValidAttribute2(lcTag, lcName, value) {
      if (FORBID_ATTR[lcName]) return false;
      if (_isPatchLinkageAttribute(lcName, lcTag)) return false;
      if (SANITIZE_DOM && (lcName === "id" || lcName === "name") && (value in document2 || value in formElement)) return false;
      const nameIsPermitted = ALLOWED_ATTR[lcName] || EXTRA_ELEMENT_HANDLING.attributeCheck instanceof Function && EXTRA_ELEMENT_HANDLING.attributeCheck(lcName, lcTag);
      if (ALLOW_DATA_ATTR && regExpTest(DATA_ATTR$1, lcName)) return true;
      if (ALLOW_ARIA_ATTR && regExpTest(ARIA_ATTR$1, lcName)) return true;
      if (!nameIsPermitted) return _isBasicCustomElement(lcTag) && _matchesNameCheck(CUSTOM_ELEMENT_HANDLING.tagNameCheck, lcTag) && _matchesNameCheck(CUSTOM_ELEMENT_HANDLING.attributeNameCheck, lcName, lcTag) || lcName === "is" && CUSTOM_ELEMENT_HANDLING.allowCustomizedBuiltInElements && _matchesNameCheck(CUSTOM_ELEMENT_HANDLING.tagNameCheck, value);
      if (URI_SAFE_ATTRIBUTES[lcName]) return true;
      const compactValue = stringReplace(value, ATTR_WHITESPACE$1, "");
      if (regExpTest(IS_ALLOWED_URI$1, compactValue)) return true;
      if ((lcName === "src" || lcName === "xlink:href" || lcName === "href") && lcTag !== "script" && stringIndexOf(value, "data:") === 0 && DATA_URI_TAGS[lcTag]) return true;
      if (ALLOW_UNKNOWN_PROTOCOLS && !regExpTest(IS_SCRIPT_OR_DATA$1, compactValue)) return true;
      return !value;
    };
    const RESERVED_CUSTOM_ELEMENT_NAMES = addToSet({}, [
      "annotation-xml",
      "color-profile",
      "font-face",
      "font-face-format",
      "font-face-name",
      "font-face-src",
      "font-face-uri",
      "missing-glyph"
    ]);
    const _isBasicCustomElement = function _isBasicCustomElement2(tagName) {
      return !RESERVED_CUSTOM_ELEMENT_NAMES[stringToLowerCase(tagName)] && regExpTest(CUSTOM_ELEMENT$1, tagName);
    };
    const _applyTrustedTypesToAttribute = function _applyTrustedTypesToAttribute2(lcTag, lcName, namespaceURI, value) {
      if (trustedTypesPolicy && typeof trustedTypes === "object" && typeof trustedTypes.getAttributeType === "function" && !namespaceURI) switch (trustedTypes.getAttributeType(lcTag, lcName)) {
        case "TrustedHTML":
          return _createTrustedHTML(value);
        case "TrustedScriptURL":
          return _createTrustedScriptURL(value);
      }
      return value;
    };
    const _setAttributeValue = function _setAttributeValue2(currentNode, name, namespaceURI, value) {
      try {
        if (namespaceURI) currentNode.setAttributeNS(namespaceURI, name, value);
        else currentNode.setAttribute(name, value);
        if (_isClobbered(currentNode)) {
          _forceRemove(currentNode);
          return false;
        }
        return true;
      } catch (_) {
        _removeAttribute(name, currentNode);
        return false;
      }
    };
    const _isUnsafeAttributeValue = function _isUnsafeAttributeValue2(name, value) {
      if (SAFE_FOR_XML && regExpTest(ATTR_VALUE_BREAKOUT, value)) return true;
      return stringToLowerCase(name) === "attributename" && stringIndexOf(value, "href") !== -1;
    };
    const _sanitizeAttributes = function _sanitizeAttributes2(currentNode, root) {
      _executeHooks(hooks.beforeSanitizeAttributes, currentNode, null);
      if (_handleHookDetachedNode(currentNode, root)) return;
      const attributes = _readAttributes(currentNode);
      if (!attributes || _isClobbered(currentNode)) return;
      ALLOWED_ATTR = _forkSharedAllowlist(hooks.uponSanitizeAttribute, ALLOWED_ATTR, DEFAULT_ALLOWED_ATTR, SET_CONFIG_ALLOWED_ATTR);
      const hookEvent = {
        attrName: "",
        attrValue: "",
        keepAttr: true,
        allowedAttributes: ALLOWED_ATTR,
        forceKeepAttr: void 0
      };
      let l = attributes.length;
      const lcTag = transformCaseFunc(_readNodeName(currentNode));
      while (l--) {
        const attr = attributes[l];
        if (!attr) continue;
        const name = attr.name, namespaceURI = attr.namespaceURI, attrValue = attr.value;
        const lcName = transformCaseFunc(name);
        const initValue = attrValue;
        let value = name === "value" ? initValue : stringTrim(initValue);
        let recreatedNamedProp = false;
        hookEvent.attrName = lcName;
        hookEvent.attrValue = value;
        hookEvent.keepAttr = true;
        hookEvent.forceKeepAttr = void 0;
        _executeHooks(hooks.uponSanitizeAttribute, currentNode, hookEvent);
        value = hookEvent.attrValue;
        if (SANITIZE_NAMED_PROPS && (lcName === "id" || lcName === "name") && stringIndexOf(value, SANITIZE_NAMED_PROPS_PREFIX) !== 0) {
          _removeAttribute(name, currentNode, attr);
          value = SANITIZE_NAMED_PROPS_PREFIX + value;
          recreatedNamedProp = true;
        }
        if (_isUnsafeAttributeValue(name, value)) {
          _removeAttribute(name, currentNode, attr);
          continue;
        }
        if (hookEvent.forceKeepAttr) continue;
        if (!hookEvent.keepAttr) {
          _removeAttribute(name, currentNode, attr);
          continue;
        }
        if (!ALLOW_SELF_CLOSE_IN_ATTR && regExpTest(SELF_CLOSING_TAG, value)) {
          _removeAttribute(name, currentNode, attr);
          continue;
        }
        if (SAFE_FOR_TEMPLATES) value = _stripTemplateExpressions(value);
        if (!_isValidAttribute(lcTag, lcName, value)) {
          _removeAttribute(name, currentNode, attr);
          continue;
        }
        value = _applyTrustedTypesToAttribute(lcTag, lcName, namespaceURI, value);
        if (value !== initValue) {
          if (_setAttributeValue(currentNode, name, namespaceURI, value) && recreatedNamedProp) arrayPop(DOMPurify.removed);
        }
      }
      _executeHooks(hooks.afterSanitizeAttributes, currentNode, null);
      _handleHookDetachedNode(currentNode, root);
    };
    const _sanitizeNode = function _sanitizeNode2(node, root) {
      _sanitizeElements(node, root);
      _sanitizeAttributes(node, root);
      if (_isDocumentFragment(node.content)) _sanitizeShadowDOM2(node.content);
    };
    const _sanitizeShadowDOM2 = function _sanitizeShadowDOM(fragment) {
      let shadowNode = null;
      const shadowIterator = _createNodeIterator(fragment);
      _executeHooks(hooks.beforeSanitizeShadowDOM, fragment, null);
      while (shadowNode = shadowIterator.nextNode()) {
        _executeHooks(hooks.uponSanitizeShadowNode, shadowNode, null);
        _sanitizeNode(shadowNode, fragment);
        if (_readNodeType(shadowNode) === NODE_TYPE.element) {
          const innerSr = getShadowRoot(shadowNode);
          if (_isDocumentFragment(innerSr)) {
            _sanitizeAttachedShadowRoots(innerSr);
            _sanitizeShadowDOM2(innerSr);
          }
        }
      }
      if (SAFE_FOR_TEMPLATES) _scrubOwnCharacterData(fragment);
      _executeHooks(hooks.afterSanitizeShadowDOM, fragment, null);
    };
    const _sanitizeAttachedShadowRoots = function _sanitizeAttachedShadowRoots2(root) {
      const stack = [{
        node: root,
        shadow: null
      }];
      while (stack.length > 0) {
        const item = stack.pop();
        if (item.shadow) {
          _sanitizeShadowDOM2(item.shadow);
          continue;
        }
        const node = item.node;
        const isElement = _readNodeType(node) === NODE_TYPE.element;
        const childNodes = getChildNodes(node);
        if (childNodes) for (let i = childNodes.length - 1; i >= 0; --i) stack.push({
          node: childNodes[i],
          shadow: null
        });
        if (isElement) {
          const rootName = getNodeName ? getNodeName(node) : null;
          if (typeof rootName === "string" && transformCaseFunc(rootName) === "template") {
            const content = node.content;
            if (_isDocumentFragment(content)) stack.push({
              node: content,
              shadow: null
            });
          }
        }
        if (isElement) {
          const sr = getShadowRoot(node);
          if (_isDocumentFragment(sr)) stack.push({
            node: null,
            shadow: sr
          }, {
            node: sr,
            shadow: null
          });
        }
      }
    };
    const _sanitize = function _sanitize2(dirty) {
      let cfg = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {};
      let body = null;
      let importedNode = null;
      let currentNode = null;
      let returnNode = null;
      IS_EMPTY_INPUT = !dirty;
      if (IS_EMPTY_INPUT) dirty = "<!-->";
      if (typeof dirty !== "string" && !_isNode(dirty)) {
        dirty = stringifyValue(dirty);
        if (typeof dirty !== "string") throw typeErrorCreate("dirty is not a string, aborting");
      }
      if (!DOMPurify.isSupported) return dirty;
      if (SET_CONFIG) {
        ALLOWED_TAGS = SET_CONFIG_ALLOWED_TAGS;
        ALLOWED_ATTR = SET_CONFIG_ALLOWED_ATTR;
      } else _parseConfig(cfg);
      ACTIVE_CONFIG = CONFIG;
      DOMPurify.removed = [];
      const inPlace = IN_PLACE && typeof dirty !== "string" && _isNode(dirty);
      let inPlaceNodes = [];
      const outerVisited = IN_PLACE_VISITED;
      const outerReturns = IN_PLACE_RETURNS;
      if (inPlace) {
        _neutralizePatchLinkage(dirty);
        inPlaceNodes = _collectLiveNodes(dirty);
        const nn = _readNodeName(dirty);
        if (typeof nn === "string") {
          const tagName = transformCaseFunc(nn);
          if (!ALLOWED_TAGS[tagName] || FORBID_TAGS[tagName]) {
            _failClosed(dirty, inPlaceNodes);
            throw typeErrorCreate("root node is forbidden and cannot be sanitized in-place");
          }
        }
        if (_isClobbered(dirty)) {
          _failClosed(dirty, inPlaceNodes);
          throw typeErrorCreate("root node is clobbered and cannot be sanitized in-place");
        }
        IN_PLACE_VISITED = setCreate();
        IN_PLACE_RETURNS = mapCreate();
      } else if (_isNode(dirty)) {
        body = _initDocument("<!---->");
        importedNode = body.ownerDocument.importNode(dirty, true);
        if (importedNode.nodeType === NODE_TYPE.element && importedNode.nodeName === "BODY") body = importedNode;
        else if (importedNode.nodeName === "HTML") body = importedNode;
        else body.appendChild(importedNode);
        _sanitizeAttachedShadowRoots(body);
      } else {
        if (!RETURN_DOM && !SAFE_FOR_TEMPLATES && !WHOLE_DOCUMENT && dirty.indexOf("<") === -1) return _signOutput(dirty);
        body = _initDocument(dirty);
        if (!body) return RETURN_DOM ? null : RETURN_TRUSTED_TYPE ? emptyHTML : "";
      }
      if (body && FORCE_BODY) _forceRemove(body.firstChild);
      const walkRoot = inPlace ? dirty : body;
      try {
        if (inPlace) _sanitizeAttachedShadowRoots(dirty);
        const nodeIterator = _createNodeIterator(walkRoot);
        while (currentNode = nodeIterator.nextNode()) _sanitizeNode(currentNode, walkRoot);
        if (inPlace) _sanitizeUnvisitedNodes(dirty, inPlaceNodes);
      } catch (error) {
        if (inPlace) _failClosed(dirty, inPlaceNodes);
        throw error;
      } finally {
        IN_PLACE_VISITED = outerVisited;
        IN_PLACE_RETURNS = outerReturns;
      }
      if (inPlace) {
        if (_neutralizeDiscarded(dirty, inPlaceNodes)) throw typeErrorCreate("a node selected for removal could not be safely returned; refusing to sanitize in place");
        if (SAFE_FOR_TEMPLATES) _scrubTemplateExpressions2(dirty);
        return dirty;
      }
      if (RETURN_DOM) {
        if (SAFE_FOR_TEMPLATES) _scrubTemplateExpressions2(body);
        if (RETURN_DOM_FRAGMENT) {
          returnNode = createDocumentFragment.call(body.ownerDocument);
          while (body.firstChild) returnNode.appendChild(body.firstChild);
        } else returnNode = body;
        if (ALLOWED_ATTR.shadowroot || ALLOWED_ATTR.shadowrootmode) returnNode = importNode.call(originalDocument, returnNode, true);
        return returnNode;
      }
      if (SAFE_FOR_TEMPLATES) _scrubOwnCharacterData(body);
      let serializedHTML = WHOLE_DOCUMENT ? body.outerHTML : body.innerHTML;
      if (WHOLE_DOCUMENT && ALLOWED_TAGS["!doctype"] && body.ownerDocument && body.ownerDocument.doctype && body.ownerDocument.doctype.name && regExpTest(DOCTYPE_NAME, body.ownerDocument.doctype.name)) serializedHTML = "<!DOCTYPE " + body.ownerDocument.doctype.name + ">\n" + serializedHTML;
      return _signOutput(serializedHTML);
    };
    DOMPurify.sanitize = function(dirty) {
      let cfg = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {};
      const outerConfig = ACTIVE_CONFIG;
      const outerTags = ALLOWED_TAGS;
      const outerAttr = ALLOWED_ATTR;
      const outerRemoved = DOMPurify.removed;
      ACTIVE_CONFIG = null;
      try {
        return _sanitize(dirty, cfg);
      } finally {
        ACTIVE_CONFIG = outerConfig;
        if (outerConfig) {
          if (!SET_CONFIG) _parseConfig(outerConfig);
          ALLOWED_TAGS = outerTags;
          ALLOWED_ATTR = outerAttr;
          DOMPurify.removed = outerRemoved;
        }
      }
    };
    DOMPurify.setConfig = function() {
      let cfg = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
      _parseConfig(cfg);
      SET_CONFIG = true;
      SET_CONFIG_ALLOWED_TAGS = ALLOWED_TAGS;
      SET_CONFIG_ALLOWED_ATTR = ALLOWED_ATTR;
    };
    DOMPurify.clearConfig = function() {
      CONFIG = null;
      SET_CONFIG = false;
      SET_CONFIG_ALLOWED_TAGS = null;
      SET_CONFIG_ALLOWED_ATTR = null;
      trustedTypesPolicy = defaultTrustedTypesPolicy;
      emptyHTML = "";
    };
    DOMPurify.isValidAttribute = function(tag2, attr, value) {
      if (!CONFIG) _parseConfig({});
      const lcTag = transformCaseFunc(tag2);
      const lcName = transformCaseFunc(attr);
      if (typeof value === "string" && (_isUnsafeAttributeValue(attr, value) || !ALLOW_SELF_CLOSE_IN_ATTR && regExpTest(SELF_CLOSING_TAG, value))) return false;
      return _isValidAttribute(lcTag, lcName, value);
    };
    DOMPurify.addHook = function(entryPoint, hookFunction) {
      if (typeof hookFunction !== "function") return;
      if (!objectHasOwnProperty(hooks, entryPoint)) return;
      arrayPush(hooks[entryPoint], hookFunction);
    };
    DOMPurify.removeHook = function(entryPoint, hookFunction) {
      if (!objectHasOwnProperty(hooks, entryPoint)) return;
      if (hookFunction !== void 0) {
        const index = arrayLastIndexOf(hooks[entryPoint], hookFunction);
        return index === -1 ? void 0 : arraySplice(hooks[entryPoint], index, 1)[0];
      }
      return arrayPop(hooks[entryPoint]);
    };
    DOMPurify.removeHooks = function(entryPoint) {
      if (!objectHasOwnProperty(hooks, entryPoint)) return;
      hooks[entryPoint] = [];
    };
    DOMPurify.removeAllHooks = function() {
      hooks = _createHooksMap();
    };
    return DOMPurify;
  }
  var purify_default = createDOMPurify();

  // ../../../web/node_modules/js-yaml/dist/js-yaml.mjs
  function getDefaultExportFromCjs(x) {
    return x && x.__esModule && Object.prototype.hasOwnProperty.call(x, "default") ? x["default"] : x;
  }
  var jsYaml = {};
  var loader = {};
  var common = {};
  var hasRequiredCommon;
  function requireCommon() {
    if (hasRequiredCommon) return common;
    hasRequiredCommon = 1;
    function isNothing(subject) {
      return typeof subject === "undefined" || subject === null;
    }
    function isObject(subject) {
      return typeof subject === "object" && subject !== null;
    }
    function toArray(sequence) {
      if (Array.isArray(sequence)) return sequence;
      else if (isNothing(sequence)) return [];
      return [sequence];
    }
    function extend(target, source) {
      if (source) {
        const sourceKeys = Object.keys(source);
        for (let index = 0, length = sourceKeys.length; index < length; index += 1) {
          const key = sourceKeys[index];
          target[key] = source[key];
        }
      }
      return target;
    }
    function repeat(string, count) {
      let result = "";
      for (let cycle = 0; cycle < count; cycle += 1) {
        result += string;
      }
      return result;
    }
    function isNegativeZero(number) {
      return number === 0 && Number.NEGATIVE_INFINITY === 1 / number;
    }
    common.isNothing = isNothing;
    common.isObject = isObject;
    common.toArray = toArray;
    common.repeat = repeat;
    common.isNegativeZero = isNegativeZero;
    common.extend = extend;
    return common;
  }
  var exception;
  var hasRequiredException;
  function requireException() {
    if (hasRequiredException) return exception;
    hasRequiredException = 1;
    function formatError(exception2, compact) {
      let where = "";
      const message = exception2.reason || "(unknown reason)";
      if (!exception2.mark) return message;
      if (exception2.mark.name) {
        where += 'in "' + exception2.mark.name + '" ';
      }
      where += "(" + (exception2.mark.line + 1) + ":" + (exception2.mark.column + 1) + ")";
      if (!compact && exception2.mark.snippet) {
        where += "\n\n" + exception2.mark.snippet;
      }
      return message + " " + where;
    }
    function YAMLException2(reason, mark) {
      Error.call(this);
      this.name = "YAMLException";
      this.reason = reason;
      this.mark = mark;
      this.message = formatError(this, false);
      if (Error.captureStackTrace) {
        Error.captureStackTrace(this, this.constructor);
      } else {
        this.stack = new Error().stack || "";
      }
    }
    YAMLException2.prototype = Object.create(Error.prototype);
    YAMLException2.prototype.constructor = YAMLException2;
    YAMLException2.prototype.toString = function toString(compact) {
      return this.name + ": " + formatError(this, compact);
    };
    exception = YAMLException2;
    return exception;
  }
  var snippet;
  var hasRequiredSnippet;
  function requireSnippet() {
    if (hasRequiredSnippet) return snippet;
    hasRequiredSnippet = 1;
    const common2 = requireCommon();
    function getLine(buffer, lineStart, lineEnd, position, maxLineLength) {
      let head = "";
      let tail = "";
      const maxHalfLength = Math.floor(maxLineLength / 2) - 1;
      if (position - lineStart > maxHalfLength) {
        head = " ... ";
        lineStart = position - maxHalfLength + head.length;
      }
      if (lineEnd - position > maxHalfLength) {
        tail = " ...";
        lineEnd = position + maxHalfLength - tail.length;
      }
      return {
        str: head + buffer.slice(lineStart, lineEnd).replace(/\t/g, "\u2192") + tail,
        pos: position - lineStart + head.length
        // relative position
      };
    }
    function padStart(string, max) {
      return common2.repeat(" ", max - string.length) + string;
    }
    function makeSnippet(mark, options2) {
      options2 = Object.create(options2 || null);
      if (!mark.buffer) return null;
      if (!options2.maxLength) options2.maxLength = 79;
      if (typeof options2.indent !== "number") options2.indent = 1;
      if (typeof options2.linesBefore !== "number") options2.linesBefore = 3;
      if (typeof options2.linesAfter !== "number") options2.linesAfter = 2;
      const re = /\r?\n|\r|\0/g;
      const lineStarts = [0];
      const lineEnds = [];
      let match;
      let foundLineNo = -1;
      while (match = re.exec(mark.buffer)) {
        lineEnds.push(match.index);
        lineStarts.push(match.index + match[0].length);
        if (mark.position <= match.index && foundLineNo < 0) {
          foundLineNo = lineStarts.length - 2;
        }
      }
      if (foundLineNo < 0) foundLineNo = lineStarts.length - 1;
      let result = "";
      const lineNoLength = Math.min(mark.line + options2.linesAfter, lineEnds.length).toString().length;
      const maxLineLength = options2.maxLength - (options2.indent + lineNoLength + 3);
      for (let i = 1; i <= options2.linesBefore; i++) {
        if (foundLineNo - i < 0) break;
        const line2 = getLine(
          mark.buffer,
          lineStarts[foundLineNo - i],
          lineEnds[foundLineNo - i],
          mark.position - (lineStarts[foundLineNo] - lineStarts[foundLineNo - i]),
          maxLineLength
        );
        result = common2.repeat(" ", options2.indent) + padStart((mark.line - i + 1).toString(), lineNoLength) + " | " + line2.str + "\n" + result;
      }
      const line = getLine(mark.buffer, lineStarts[foundLineNo], lineEnds[foundLineNo], mark.position, maxLineLength);
      result += common2.repeat(" ", options2.indent) + padStart((mark.line + 1).toString(), lineNoLength) + " | " + line.str + "\n";
      result += common2.repeat("-", options2.indent + lineNoLength + 3 + line.pos) + "^\n";
      for (let i = 1; i <= options2.linesAfter; i++) {
        if (foundLineNo + i >= lineEnds.length) break;
        const line2 = getLine(
          mark.buffer,
          lineStarts[foundLineNo + i],
          lineEnds[foundLineNo + i],
          mark.position - (lineStarts[foundLineNo] - lineStarts[foundLineNo + i]),
          maxLineLength
        );
        result += common2.repeat(" ", options2.indent) + padStart((mark.line + i + 1).toString(), lineNoLength) + " | " + line2.str + "\n";
      }
      return result.replace(/\n$/, "");
    }
    snippet = makeSnippet;
    return snippet;
  }
  var type;
  var hasRequiredType;
  function requireType() {
    if (hasRequiredType) return type;
    hasRequiredType = 1;
    const YAMLException2 = requireException();
    const TYPE_CONSTRUCTOR_OPTIONS = [
      "kind",
      "multi",
      "resolve",
      "construct",
      "instanceOf",
      "predicate",
      "represent",
      "representName",
      "defaultStyle",
      "styleAliases"
    ];
    const YAML_NODE_KINDS = [
      "scalar",
      "sequence",
      "mapping"
    ];
    function compileStyleAliases(map2) {
      const result = {};
      if (map2 !== null) {
        Object.keys(map2).forEach(function(style) {
          map2[style].forEach(function(alias) {
            result[String(alias)] = style;
          });
        });
      }
      return result;
    }
    function Type2(tag2, options2) {
      options2 = options2 || {};
      Object.keys(options2).forEach(function(name) {
        if (TYPE_CONSTRUCTOR_OPTIONS.indexOf(name) === -1) {
          throw new YAMLException2('Unknown option "' + name + '" is met in definition of "' + tag2 + '" YAML type.');
        }
      });
      this.options = options2;
      this.tag = tag2;
      this.kind = options2["kind"] || null;
      this.resolve = options2["resolve"] || function() {
        return true;
      };
      this.construct = options2["construct"] || function(data) {
        return data;
      };
      this.instanceOf = options2["instanceOf"] || null;
      this.predicate = options2["predicate"] || null;
      this.represent = options2["represent"] || null;
      this.representName = options2["representName"] || null;
      this.defaultStyle = options2["defaultStyle"] || null;
      this.multi = options2["multi"] || false;
      this.styleAliases = compileStyleAliases(options2["styleAliases"] || null);
      if (YAML_NODE_KINDS.indexOf(this.kind) === -1) {
        throw new YAMLException2('Unknown kind "' + this.kind + '" is specified for "' + tag2 + '" YAML type.');
      }
    }
    type = Type2;
    return type;
  }
  var schema;
  var hasRequiredSchema;
  function requireSchema() {
    if (hasRequiredSchema) return schema;
    hasRequiredSchema = 1;
    const YAMLException2 = requireException();
    const Type2 = requireType();
    function compileList(schema2, name) {
      const result = [];
      schema2[name].forEach(function(currentType) {
        let newIndex = result.length;
        result.forEach(function(previousType, previousIndex) {
          if (previousType.tag === currentType.tag && previousType.kind === currentType.kind && previousType.multi === currentType.multi) {
            newIndex = previousIndex;
          }
        });
        result[newIndex] = currentType;
      });
      return result;
    }
    function compileMap() {
      const result = {
        scalar: {},
        sequence: {},
        mapping: {},
        fallback: {},
        multi: {
          scalar: [],
          sequence: [],
          mapping: [],
          fallback: []
        }
      };
      function collectType(type2) {
        if (type2.multi) {
          result.multi[type2.kind].push(type2);
          result.multi["fallback"].push(type2);
        } else {
          result[type2.kind][type2.tag] = result["fallback"][type2.tag] = type2;
        }
      }
      for (let index = 0, length = arguments.length; index < length; index += 1) {
        arguments[index].forEach(collectType);
      }
      return result;
    }
    function Schema2(definition) {
      return this.extend(definition);
    }
    Schema2.prototype.extend = function extend(definition) {
      let implicit = [];
      let explicit = [];
      if (definition instanceof Type2) {
        explicit.push(definition);
      } else if (Array.isArray(definition)) {
        explicit = explicit.concat(definition);
      } else if (definition && (Array.isArray(definition.implicit) || Array.isArray(definition.explicit))) {
        if (definition.implicit) implicit = implicit.concat(definition.implicit);
        if (definition.explicit) explicit = explicit.concat(definition.explicit);
      } else {
        throw new YAMLException2("Schema.extend argument should be a Type, [ Type ], or a schema definition ({ implicit: [...], explicit: [...] })");
      }
      implicit.forEach(function(type2) {
        if (!(type2 instanceof Type2)) {
          throw new YAMLException2("Specified list of YAML types (or a single Type object) contains a non-Type object.");
        }
        if (type2.loadKind && type2.loadKind !== "scalar") {
          throw new YAMLException2("There is a non-scalar type in the implicit list of a schema. Implicit resolving of such types is not supported.");
        }
        if (type2.multi) {
          throw new YAMLException2("There is a multi type in the implicit list of a schema. Multi tags can only be listed as explicit.");
        }
      });
      explicit.forEach(function(type2) {
        if (!(type2 instanceof Type2)) {
          throw new YAMLException2("Specified list of YAML types (or a single Type object) contains a non-Type object.");
        }
      });
      const result = Object.create(Schema2.prototype);
      result.implicit = (this.implicit || []).concat(implicit);
      result.explicit = (this.explicit || []).concat(explicit);
      result.compiledImplicit = compileList(result, "implicit");
      result.compiledExplicit = compileList(result, "explicit");
      result.compiledTypeMap = compileMap(result.compiledImplicit, result.compiledExplicit);
      return result;
    };
    schema = Schema2;
    return schema;
  }
  var str;
  var hasRequiredStr;
  function requireStr() {
    if (hasRequiredStr) return str;
    hasRequiredStr = 1;
    const Type2 = requireType();
    str = new Type2("tag:yaml.org,2002:str", {
      kind: "scalar",
      construct: function(data) {
        return data !== null ? data : "";
      }
    });
    return str;
  }
  var seq;
  var hasRequiredSeq;
  function requireSeq() {
    if (hasRequiredSeq) return seq;
    hasRequiredSeq = 1;
    const Type2 = requireType();
    seq = new Type2("tag:yaml.org,2002:seq", {
      kind: "sequence",
      construct: function(data) {
        return data !== null ? data : [];
      }
    });
    return seq;
  }
  var map;
  var hasRequiredMap;
  function requireMap() {
    if (hasRequiredMap) return map;
    hasRequiredMap = 1;
    const Type2 = requireType();
    map = new Type2("tag:yaml.org,2002:map", {
      kind: "mapping",
      construct: function(data) {
        return data !== null ? data : {};
      }
    });
    return map;
  }
  var failsafe;
  var hasRequiredFailsafe;
  function requireFailsafe() {
    if (hasRequiredFailsafe) return failsafe;
    hasRequiredFailsafe = 1;
    const Schema2 = requireSchema();
    failsafe = new Schema2({
      explicit: [
        requireStr(),
        requireSeq(),
        requireMap()
      ]
    });
    return failsafe;
  }
  var _null;
  var hasRequired_null;
  function require_null() {
    if (hasRequired_null) return _null;
    hasRequired_null = 1;
    const Type2 = requireType();
    function resolveYamlNull(data) {
      if (data === null) return true;
      const max = data.length;
      return max === 1 && data === "~" || max === 4 && (data === "null" || data === "Null" || data === "NULL");
    }
    function constructYamlNull() {
      return null;
    }
    function isNull(object) {
      return object === null;
    }
    _null = new Type2("tag:yaml.org,2002:null", {
      kind: "scalar",
      resolve: resolveYamlNull,
      construct: constructYamlNull,
      predicate: isNull,
      represent: {
        canonical: function() {
          return "~";
        },
        lowercase: function() {
          return "null";
        },
        uppercase: function() {
          return "NULL";
        },
        camelcase: function() {
          return "Null";
        },
        empty: function() {
          return "";
        }
      },
      defaultStyle: "lowercase"
    });
    return _null;
  }
  var bool;
  var hasRequiredBool;
  function requireBool() {
    if (hasRequiredBool) return bool;
    hasRequiredBool = 1;
    const Type2 = requireType();
    function resolveYamlBoolean(data) {
      if (data === null) return false;
      const max = data.length;
      return max === 4 && (data === "true" || data === "True" || data === "TRUE") || max === 5 && (data === "false" || data === "False" || data === "FALSE");
    }
    function constructYamlBoolean(data) {
      return data === "true" || data === "True" || data === "TRUE";
    }
    function isBoolean(object) {
      return Object.prototype.toString.call(object) === "[object Boolean]";
    }
    bool = new Type2("tag:yaml.org,2002:bool", {
      kind: "scalar",
      resolve: resolveYamlBoolean,
      construct: constructYamlBoolean,
      predicate: isBoolean,
      represent: {
        lowercase: function(object) {
          return object ? "true" : "false";
        },
        uppercase: function(object) {
          return object ? "TRUE" : "FALSE";
        },
        camelcase: function(object) {
          return object ? "True" : "False";
        }
      },
      defaultStyle: "lowercase"
    });
    return bool;
  }
  var int;
  var hasRequiredInt;
  function requireInt() {
    if (hasRequiredInt) return int;
    hasRequiredInt = 1;
    const common2 = requireCommon();
    const Type2 = requireType();
    function isHexCode(c) {
      return c >= 48 && c <= 57 || c >= 65 && c <= 70 || c >= 97 && c <= 102;
    }
    function isOctCode(c) {
      return c >= 48 && c <= 55;
    }
    function isDecCode(c) {
      return c >= 48 && c <= 57;
    }
    function resolveYamlInteger(data) {
      if (data === null) return false;
      const max = data.length;
      let index = 0;
      let hasDigits = false;
      if (!max) return false;
      let ch = data[index];
      if (ch === "-" || ch === "+") {
        ch = data[++index];
      }
      if (ch === "0") {
        if (index + 1 === max) return true;
        ch = data[++index];
        if (ch === "b") {
          index++;
          for (; index < max; index++) {
            ch = data[index];
            if (ch !== "0" && ch !== "1") return false;
            hasDigits = true;
          }
          return hasDigits && isFinite(parseYamlInteger(data));
        }
        if (ch === "x") {
          index++;
          for (; index < max; index++) {
            if (!isHexCode(data.charCodeAt(index))) return false;
            hasDigits = true;
          }
          return hasDigits && isFinite(parseYamlInteger(data));
        }
        if (ch === "o") {
          index++;
          for (; index < max; index++) {
            if (!isOctCode(data.charCodeAt(index))) return false;
            hasDigits = true;
          }
          return hasDigits && isFinite(parseYamlInteger(data));
        }
      }
      for (; index < max; index++) {
        if (!isDecCode(data.charCodeAt(index))) {
          return false;
        }
        hasDigits = true;
      }
      if (!hasDigits) return false;
      return isFinite(parseYamlInteger(data));
    }
    function parseYamlInteger(data) {
      let value = data;
      let sign = 1;
      let ch = value[0];
      if (ch === "-" || ch === "+") {
        if (ch === "-") sign = -1;
        value = value.slice(1);
        ch = value[0];
      }
      if (value === "0") return 0;
      if (ch === "0") {
        if (value[1] === "b") return sign * parseInt(value.slice(2), 2);
        if (value[1] === "x") return sign * parseInt(value.slice(2), 16);
        if (value[1] === "o") return sign * parseInt(value.slice(2), 8);
      }
      return sign * parseInt(value, 10);
    }
    function constructYamlInteger(data) {
      return parseYamlInteger(data);
    }
    function isInteger2(object) {
      return Object.prototype.toString.call(object) === "[object Number]" && (object % 1 === 0 && !common2.isNegativeZero(object));
    }
    int = new Type2("tag:yaml.org,2002:int", {
      kind: "scalar",
      resolve: resolveYamlInteger,
      construct: constructYamlInteger,
      predicate: isInteger2,
      represent: {
        binary: function(obj) {
          return obj >= 0 ? "0b" + obj.toString(2) : "-0b" + obj.toString(2).slice(1);
        },
        octal: function(obj) {
          return obj >= 0 ? "0o" + obj.toString(8) : "-0o" + obj.toString(8).slice(1);
        },
        decimal: function(obj) {
          return obj.toString(10);
        },
        hexadecimal: function(obj) {
          return obj >= 0 ? "0x" + obj.toString(16).toUpperCase() : "-0x" + obj.toString(16).toUpperCase().slice(1);
        }
      },
      defaultStyle: "decimal",
      styleAliases: {
        binary: [2, "bin"],
        octal: [8, "oct"],
        decimal: [10, "dec"],
        hexadecimal: [16, "hex"]
      }
    });
    return int;
  }
  var float;
  var hasRequiredFloat;
  function requireFloat() {
    if (hasRequiredFloat) return float;
    hasRequiredFloat = 1;
    const common2 = requireCommon();
    const Type2 = requireType();
    const YAML_FLOAT_PATTERN = new RegExp(
      // 2.5e4, 2.5 and integers
      "^(?:[-+]?(?:[0-9]+)(?:\\.[0-9]*)?(?:[eE][-+]?[0-9]+)?|\\.[0-9]+(?:[eE][-+]?[0-9]+)?|[-+]?\\.(?:inf|Inf|INF)|\\.(?:nan|NaN|NAN))$"
    );
    const YAML_FLOAT_SPECIAL_PATTERN = new RegExp(
      "^(?:[-+]?\\.(?:inf|Inf|INF)|\\.(?:nan|NaN|NAN))$"
    );
    function resolveYamlFloat(data) {
      if (data === null) return false;
      if (!YAML_FLOAT_PATTERN.test(data)) {
        return false;
      }
      if (isFinite(parseFloat(data, 10))) {
        return true;
      }
      return YAML_FLOAT_SPECIAL_PATTERN.test(data);
    }
    function constructYamlFloat(data) {
      let value = data.toLowerCase();
      const sign = value[0] === "-" ? -1 : 1;
      if ("+-".indexOf(value[0]) >= 0) {
        value = value.slice(1);
      }
      if (value === ".inf") {
        return sign === 1 ? Number.POSITIVE_INFINITY : Number.NEGATIVE_INFINITY;
      } else if (value === ".nan") {
        return NaN;
      }
      return sign * parseFloat(value, 10);
    }
    const SCIENTIFIC_WITHOUT_DOT = /^[-+]?[0-9]+e/;
    function representYamlFloat(object, style) {
      if (isNaN(object)) {
        switch (style) {
          case "lowercase":
            return ".nan";
          case "uppercase":
            return ".NAN";
          case "camelcase":
            return ".NaN";
        }
      } else if (Number.POSITIVE_INFINITY === object) {
        switch (style) {
          case "lowercase":
            return ".inf";
          case "uppercase":
            return ".INF";
          case "camelcase":
            return ".Inf";
        }
      } else if (Number.NEGATIVE_INFINITY === object) {
        switch (style) {
          case "lowercase":
            return "-.inf";
          case "uppercase":
            return "-.INF";
          case "camelcase":
            return "-.Inf";
        }
      } else if (common2.isNegativeZero(object)) {
        return "-0.0";
      }
      const res = object.toString(10);
      return SCIENTIFIC_WITHOUT_DOT.test(res) ? res.replace("e", ".e") : res;
    }
    function isFloat(object) {
      return Object.prototype.toString.call(object) === "[object Number]" && (object % 1 !== 0 || common2.isNegativeZero(object));
    }
    float = new Type2("tag:yaml.org,2002:float", {
      kind: "scalar",
      resolve: resolveYamlFloat,
      construct: constructYamlFloat,
      predicate: isFloat,
      represent: representYamlFloat,
      defaultStyle: "lowercase"
    });
    return float;
  }
  var json;
  var hasRequiredJson;
  function requireJson() {
    if (hasRequiredJson) return json;
    hasRequiredJson = 1;
    json = requireFailsafe().extend({
      implicit: [
        require_null(),
        requireBool(),
        requireInt(),
        requireFloat()
      ]
    });
    return json;
  }
  var core;
  var hasRequiredCore;
  function requireCore() {
    if (hasRequiredCore) return core;
    hasRequiredCore = 1;
    core = requireJson();
    return core;
  }
  var timestamp;
  var hasRequiredTimestamp;
  function requireTimestamp() {
    if (hasRequiredTimestamp) return timestamp;
    hasRequiredTimestamp = 1;
    const Type2 = requireType();
    const YAML_DATE_REGEXP = new RegExp(
      "^([0-9][0-9][0-9][0-9])-([0-9][0-9])-([0-9][0-9])$"
    );
    const YAML_TIMESTAMP_REGEXP = new RegExp(
      "^([0-9][0-9][0-9][0-9])-([0-9][0-9]?)-([0-9][0-9]?)(?:[Tt]|[ \\t]+)([0-9][0-9]?):([0-9][0-9]):([0-9][0-9])(?:\\.([0-9]*))?(?:[ \\t]*(Z|([-+])([0-9][0-9]?)(?::([0-9][0-9]))?))?$"
    );
    function resolveYamlTimestamp(data) {
      if (data === null) return false;
      if (YAML_DATE_REGEXP.exec(data) !== null) return true;
      if (YAML_TIMESTAMP_REGEXP.exec(data) !== null) return true;
      return false;
    }
    function constructYamlTimestamp(data) {
      let fraction = 0;
      let delta = null;
      let match = YAML_DATE_REGEXP.exec(data);
      if (match === null) match = YAML_TIMESTAMP_REGEXP.exec(data);
      if (match === null) throw new Error("Date resolve error");
      const year = +match[1];
      const month = +match[2] - 1;
      const day = +match[3];
      if (!match[4]) {
        return new Date(Date.UTC(year, month, day));
      }
      const hour = +match[4];
      const minute = +match[5];
      const second = +match[6];
      if (match[7]) {
        fraction = match[7].slice(0, 3);
        while (fraction.length < 3) {
          fraction += "0";
        }
        fraction = +fraction;
      }
      if (match[9]) {
        const tzHour = +match[10];
        const tzMinute = +(match[11] || 0);
        delta = (tzHour * 60 + tzMinute) * 6e4;
        if (match[9] === "-") delta = -delta;
      }
      const date = new Date(Date.UTC(year, month, day, hour, minute, second, fraction));
      if (delta) date.setTime(date.getTime() - delta);
      return date;
    }
    function representYamlTimestamp(object) {
      return object.toISOString();
    }
    timestamp = new Type2("tag:yaml.org,2002:timestamp", {
      kind: "scalar",
      resolve: resolveYamlTimestamp,
      construct: constructYamlTimestamp,
      instanceOf: Date,
      represent: representYamlTimestamp
    });
    return timestamp;
  }
  var merge;
  var hasRequiredMerge;
  function requireMerge() {
    if (hasRequiredMerge) return merge;
    hasRequiredMerge = 1;
    const Type2 = requireType();
    function resolveYamlMerge(data) {
      return data === "<<" || data === null;
    }
    merge = new Type2("tag:yaml.org,2002:merge", {
      kind: "scalar",
      resolve: resolveYamlMerge
    });
    return merge;
  }
  var binary;
  var hasRequiredBinary;
  function requireBinary() {
    if (hasRequiredBinary) return binary;
    hasRequiredBinary = 1;
    const Type2 = requireType();
    const BASE64_MAP = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=\n\r";
    function resolveYamlBinary(data) {
      if (data === null) return false;
      let bitlen = 0;
      const max = data.length;
      const map2 = BASE64_MAP;
      for (let idx = 0; idx < max; idx++) {
        const code = map2.indexOf(data.charAt(idx));
        if (code > 64) continue;
        if (code < 0) return false;
        bitlen += 6;
      }
      return bitlen % 8 === 0;
    }
    function constructYamlBinary(data) {
      const input = data.replace(/[\r\n=]/g, "");
      const max = input.length;
      const map2 = BASE64_MAP;
      let bits = 0;
      const result = [];
      for (let idx = 0; idx < max; idx++) {
        if (idx % 4 === 0 && idx) {
          result.push(bits >> 16 & 255);
          result.push(bits >> 8 & 255);
          result.push(bits & 255);
        }
        bits = bits << 6 | map2.indexOf(input.charAt(idx));
      }
      const tailbits = max % 4 * 6;
      if (tailbits === 0) {
        result.push(bits >> 16 & 255);
        result.push(bits >> 8 & 255);
        result.push(bits & 255);
      } else if (tailbits === 18) {
        result.push(bits >> 10 & 255);
        result.push(bits >> 2 & 255);
      } else if (tailbits === 12) {
        result.push(bits >> 4 & 255);
      }
      return new Uint8Array(result);
    }
    function representYamlBinary(object) {
      let result = "";
      let bits = 0;
      const max = object.length;
      const map2 = BASE64_MAP;
      for (let idx = 0; idx < max; idx++) {
        if (idx % 3 === 0 && idx) {
          result += map2[bits >> 18 & 63];
          result += map2[bits >> 12 & 63];
          result += map2[bits >> 6 & 63];
          result += map2[bits & 63];
        }
        bits = (bits << 8) + object[idx];
      }
      const tail = max % 3;
      if (tail === 0) {
        result += map2[bits >> 18 & 63];
        result += map2[bits >> 12 & 63];
        result += map2[bits >> 6 & 63];
        result += map2[bits & 63];
      } else if (tail === 2) {
        result += map2[bits >> 10 & 63];
        result += map2[bits >> 4 & 63];
        result += map2[bits << 2 & 63];
        result += map2[64];
      } else if (tail === 1) {
        result += map2[bits >> 2 & 63];
        result += map2[bits << 4 & 63];
        result += map2[64];
        result += map2[64];
      }
      return result;
    }
    function isBinary(obj) {
      return Object.prototype.toString.call(obj) === "[object Uint8Array]";
    }
    binary = new Type2("tag:yaml.org,2002:binary", {
      kind: "scalar",
      resolve: resolveYamlBinary,
      construct: constructYamlBinary,
      predicate: isBinary,
      represent: representYamlBinary
    });
    return binary;
  }
  var omap;
  var hasRequiredOmap;
  function requireOmap() {
    if (hasRequiredOmap) return omap;
    hasRequiredOmap = 1;
    const Type2 = requireType();
    const _hasOwnProperty2 = Object.prototype.hasOwnProperty;
    const _toString = Object.prototype.toString;
    function resolveYamlOmap(data) {
      if (data === null) return true;
      const objectKeys2 = {};
      const object = data;
      for (let index = 0, length = object.length; index < length; index += 1) {
        const pair = object[index];
        let pairHasKey = false;
        if (_toString.call(pair) !== "[object Object]") return false;
        let pairKey;
        for (pairKey in pair) {
          if (_hasOwnProperty2.call(pair, pairKey)) {
            if (!pairHasKey) pairHasKey = true;
            else return false;
          }
        }
        if (!pairHasKey) return false;
        if (_hasOwnProperty2.call(objectKeys2, pairKey)) return false;
        Object.defineProperty(objectKeys2, pairKey, { value: true });
      }
      return true;
    }
    function constructYamlOmap(data) {
      return data !== null ? data : [];
    }
    omap = new Type2("tag:yaml.org,2002:omap", {
      kind: "sequence",
      resolve: resolveYamlOmap,
      construct: constructYamlOmap
    });
    return omap;
  }
  var pairs;
  var hasRequiredPairs;
  function requirePairs() {
    if (hasRequiredPairs) return pairs;
    hasRequiredPairs = 1;
    const Type2 = requireType();
    const _toString = Object.prototype.toString;
    function resolveYamlPairs(data) {
      if (data === null) return true;
      const object = data;
      const result = new Array(object.length);
      for (let index = 0, length = object.length; index < length; index += 1) {
        const pair = object[index];
        if (_toString.call(pair) !== "[object Object]") return false;
        const keys = Object.keys(pair);
        if (keys.length !== 1) return false;
        result[index] = [keys[0], pair[keys[0]]];
      }
      return true;
    }
    function constructYamlPairs(data) {
      if (data === null) return [];
      const object = data;
      const result = new Array(object.length);
      for (let index = 0, length = object.length; index < length; index += 1) {
        const pair = object[index];
        const keys = Object.keys(pair);
        result[index] = [keys[0], pair[keys[0]]];
      }
      return result;
    }
    pairs = new Type2("tag:yaml.org,2002:pairs", {
      kind: "sequence",
      resolve: resolveYamlPairs,
      construct: constructYamlPairs
    });
    return pairs;
  }
  var set;
  var hasRequiredSet;
  function requireSet() {
    if (hasRequiredSet) return set;
    hasRequiredSet = 1;
    const Type2 = requireType();
    const _hasOwnProperty2 = Object.prototype.hasOwnProperty;
    function resolveYamlSet(data) {
      if (data === null) return true;
      const object = data;
      for (const key in object) {
        if (_hasOwnProperty2.call(object, key)) {
          if (object[key] !== null) return false;
        }
      }
      return true;
    }
    function constructYamlSet(data) {
      return data !== null ? data : {};
    }
    set = new Type2("tag:yaml.org,2002:set", {
      kind: "mapping",
      resolve: resolveYamlSet,
      construct: constructYamlSet
    });
    return set;
  }
  var _default;
  var hasRequired_default;
  function require_default() {
    if (hasRequired_default) return _default;
    hasRequired_default = 1;
    _default = requireCore().extend({
      implicit: [
        requireTimestamp(),
        requireMerge()
      ],
      explicit: [
        requireBinary(),
        requireOmap(),
        requirePairs(),
        requireSet()
      ]
    });
    return _default;
  }
  var hasRequiredLoader;
  function requireLoader() {
    if (hasRequiredLoader) return loader;
    hasRequiredLoader = 1;
    const common2 = requireCommon();
    const YAMLException2 = requireException();
    const makeSnippet = requireSnippet();
    const DEFAULT_SCHEMA2 = require_default();
    const _hasOwnProperty2 = Object.prototype.hasOwnProperty;
    const CONTEXT_FLOW_IN = 1;
    const CONTEXT_FLOW_OUT = 2;
    const CONTEXT_BLOCK_IN = 3;
    const CONTEXT_BLOCK_OUT = 4;
    const CHOMPING_CLIP = 1;
    const CHOMPING_STRIP = 2;
    const CHOMPING_KEEP = 3;
    const PATTERN_NON_PRINTABLE = /[\x00-\x08\x0B\x0C\x0E-\x1F\x7F-\x84\x86-\x9F\uFFFE\uFFFF]|[\uD800-\uDBFF](?![\uDC00-\uDFFF])|(?:[^\uD800-\uDBFF]|^)[\uDC00-\uDFFF]/;
    const PATTERN_NON_ASCII_LINE_BREAKS = /[\x85\u2028\u2029]/;
    const PATTERN_FLOW_INDICATORS = /[,\[\]{}]/;
    const PATTERN_TAG_HANDLE = /^(?:!|!!|![0-9A-Za-z-]+!)$/;
    const PATTERN_TAG_URI = /^(?:!|[^,\[\]{}])(?:%[0-9a-f]{2}|[0-9a-z\-#;/?:@&=+$,_.!~*'()\[\]])*$/i;
    function _class(obj) {
      return Object.prototype.toString.call(obj);
    }
    function isEol(c) {
      return c === 10 || c === 13;
    }
    function isWhiteSpace(c) {
      return c === 9 || c === 32;
    }
    function isWsOrEol(c) {
      return c === 9 || c === 32 || c === 10 || c === 13;
    }
    function isFlowIndicator(c) {
      return c === 44 || c === 91 || c === 93 || c === 123 || c === 125;
    }
    function fromHexCode(c) {
      if (c >= 48 && c <= 57) {
        return c - 48;
      }
      const lc = c | 32;
      if (lc >= 97 && lc <= 102) {
        return lc - 97 + 10;
      }
      return -1;
    }
    function escapedHexLen(c) {
      if (c === 120) {
        return 2;
      }
      if (c === 117) {
        return 4;
      }
      if (c === 85) {
        return 8;
      }
      return 0;
    }
    function fromDecimalCode(c) {
      if (c >= 48 && c <= 57) {
        return c - 48;
      }
      return -1;
    }
    function simpleEscapeSequence(c) {
      switch (c) {
        case 48:
          return "\0";
        case 97:
          return "\x07";
        case 98:
          return "\b";
        case 116:
          return "	";
        case 9:
          return "	";
        case 110:
          return "\n";
        case 118:
          return "\v";
        case 102:
          return "\f";
        case 114:
          return "\r";
        case 101:
          return "\x1B";
        case 32:
          return " ";
        case 34:
          return '"';
        case 47:
          return "/";
        case 92:
          return "\\";
        case 78:
          return "\x85";
        case 95:
          return "\xA0";
        case 76:
          return "\u2028";
        case 80:
          return "\u2029";
        default:
          return "";
      }
    }
    function charFromCodepoint(c) {
      if (c <= 65535) {
        return String.fromCharCode(c);
      }
      return String.fromCharCode(
        (c - 65536 >> 10) + 55296,
        (c - 65536 & 1023) + 56320
      );
    }
    function setProperty(object, key, value) {
      if (key === "__proto__") {
        Object.defineProperty(object, key, {
          configurable: true,
          enumerable: true,
          writable: true,
          value
        });
      } else {
        object[key] = value;
      }
    }
    const simpleEscapeCheck = new Array(256);
    const simpleEscapeMap = new Array(256);
    for (let i = 0; i < 256; i++) {
      simpleEscapeCheck[i] = simpleEscapeSequence(i) ? 1 : 0;
      simpleEscapeMap[i] = simpleEscapeSequence(i);
    }
    function State(input, options2) {
      this.input = input;
      this.filename = options2["filename"] || null;
      this.schema = options2["schema"] || DEFAULT_SCHEMA2;
      this.onWarning = options2["onWarning"] || null;
      this.legacy = options2["legacy"] || false;
      this.json = options2["json"] || false;
      this.listener = options2["listener"] || null;
      this.maxDepth = typeof options2["maxDepth"] === "number" ? options2["maxDepth"] : 100;
      this.maxTotalMergeKeys = typeof options2["maxTotalMergeKeys"] === "number" ? options2["maxTotalMergeKeys"] : 1e4;
      this.implicitTypes = this.schema.compiledImplicit;
      this.typeMap = this.schema.compiledTypeMap;
      this.length = input.length;
      this.position = 0;
      this.line = 0;
      this.lineStart = 0;
      this.lineIndent = 0;
      this.depth = 0;
      this.totalMergeKeys = 0;
      this.firstTabInLine = -1;
      this.documents = [];
      this.anchorMapTransactions = [];
    }
    function generateError(state, message) {
      const mark = {
        name: state.filename,
        buffer: state.input.slice(0, -1),
        // omit trailing \0
        position: state.position,
        line: state.line,
        column: state.position - state.lineStart
      };
      mark.snippet = makeSnippet(mark);
      return new YAMLException2(message, mark);
    }
    function throwError(state, message) {
      throw generateError(state, message);
    }
    function throwWarning(state, message) {
      if (state.onWarning) {
        state.onWarning.call(null, generateError(state, message));
      }
    }
    function storeAnchor(state, name, value) {
      const transactions = state.anchorMapTransactions;
      if (transactions.length !== 0) {
        const transaction = transactions[transactions.length - 1];
        if (!_hasOwnProperty2.call(transaction, name)) {
          transaction[name] = {
            existed: _hasOwnProperty2.call(state.anchorMap, name),
            value: state.anchorMap[name]
          };
        }
      }
      state.anchorMap[name] = value;
    }
    function beginAnchorTransaction(state) {
      state.anchorMapTransactions.push(/* @__PURE__ */ Object.create(null));
    }
    function commitAnchorTransaction(state) {
      const transaction = state.anchorMapTransactions.pop();
      const transactions = state.anchorMapTransactions;
      if (transactions.length === 0) return;
      const parent = transactions[transactions.length - 1];
      const names = Object.keys(transaction);
      for (let index = 0, length = names.length; index < length; index += 1) {
        const name = names[index];
        if (!_hasOwnProperty2.call(parent, name)) {
          parent[name] = transaction[name];
        }
      }
    }
    function rollbackAnchorTransaction(state) {
      const transaction = state.anchorMapTransactions.pop();
      const names = Object.keys(transaction);
      for (let index = names.length - 1; index >= 0; index -= 1) {
        const entry = transaction[names[index]];
        if (entry.existed) {
          state.anchorMap[names[index]] = entry.value;
        } else {
          delete state.anchorMap[names[index]];
        }
      }
    }
    function snapshotState(state) {
      return {
        position: state.position,
        line: state.line,
        lineStart: state.lineStart,
        lineIndent: state.lineIndent,
        firstTabInLine: state.firstTabInLine,
        tag: state.tag,
        anchor: state.anchor,
        kind: state.kind,
        result: state.result
      };
    }
    function restoreState(state, snapshot) {
      state.position = snapshot.position;
      state.line = snapshot.line;
      state.lineStart = snapshot.lineStart;
      state.lineIndent = snapshot.lineIndent;
      state.firstTabInLine = snapshot.firstTabInLine;
      state.tag = snapshot.tag;
      state.anchor = snapshot.anchor;
      state.kind = snapshot.kind;
      state.result = snapshot.result;
    }
    const directiveHandlers = {
      YAML: function handleYamlDirective(state, name, args) {
        if (state.version !== null) {
          throwError(state, "duplication of %YAML directive");
        }
        if (args.length !== 1) {
          throwError(state, "YAML directive accepts exactly one argument");
        }
        const match = /^([0-9]+)\.([0-9]+)$/.exec(args[0]);
        if (match === null) {
          throwError(state, "ill-formed argument of the YAML directive");
        }
        const major = parseInt(match[1], 10);
        const minor = parseInt(match[2], 10);
        if (major !== 1) {
          throwError(state, "unacceptable YAML version of the document");
        }
        state.version = args[0];
        state.checkLineBreaks = minor < 2;
        if (minor !== 1 && minor !== 2) {
          throwWarning(state, "unsupported YAML version of the document");
        }
      },
      TAG: function handleTagDirective(state, name, args) {
        let prefix;
        if (args.length !== 2) {
          throwError(state, "TAG directive accepts exactly two arguments");
        }
        const handle = args[0];
        prefix = args[1];
        if (!PATTERN_TAG_HANDLE.test(handle)) {
          throwError(state, "ill-formed tag handle (first argument) of the TAG directive");
        }
        if (_hasOwnProperty2.call(state.tagMap, handle)) {
          throwError(state, 'there is a previously declared suffix for "' + handle + '" tag handle');
        }
        if (!PATTERN_TAG_URI.test(prefix)) {
          throwError(state, "ill-formed tag prefix (second argument) of the TAG directive");
        }
        try {
          prefix = decodeURIComponent(prefix);
        } catch (err) {
          throwError(state, "tag prefix is malformed: " + prefix);
        }
        state.tagMap[handle] = prefix;
      }
    };
    function captureSegment(state, start, end, checkJson) {
      if (start < end) {
        const _result = state.input.slice(start, end);
        if (checkJson) {
          for (let _position = 0, _length = _result.length; _position < _length; _position += 1) {
            const _character = _result.charCodeAt(_position);
            if (!(_character === 9 || _character >= 32 && _character <= 1114111)) {
              throwError(state, "expected valid JSON character");
            }
          }
        } else if (PATTERN_NON_PRINTABLE.test(_result)) {
          throwError(state, "the stream contains non-printable characters");
        }
        state.result += _result;
      }
    }
    function chargeMergeWork(state) {
      state.totalMergeKeys++;
      if (state.maxTotalMergeKeys !== -1 && state.totalMergeKeys > state.maxTotalMergeKeys) {
        throwError(state, "merge keys exceeded maxTotalMergeKeys (" + state.maxTotalMergeKeys + ")");
      }
    }
    function mergeMappings(state, destination, source, overridableKeys) {
      if (!common2.isObject(source)) {
        throwError(state, "cannot merge mappings; the provided source object is unacceptable");
      }
      chargeMergeWork(state);
      const sourceKeys = Object.keys(source);
      for (let index = 0, quantity = sourceKeys.length; index < quantity; index += 1) {
        const key = sourceKeys[index];
        chargeMergeWork(state);
        if (!_hasOwnProperty2.call(destination, key)) {
          setProperty(destination, key, source[key]);
          overridableKeys[key] = true;
        }
      }
    }
    function storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, valueNode, startLine, startLineStart, startPos) {
      if (Array.isArray(keyNode)) {
        keyNode = Array.prototype.slice.call(keyNode);
        for (let index = 0, quantity = keyNode.length; index < quantity; index += 1) {
          if (Array.isArray(keyNode[index])) {
            throwError(state, "nested arrays are not supported inside keys");
          }
          if (typeof keyNode === "object" && _class(keyNode[index]) === "[object Object]") {
            keyNode[index] = "[object Object]";
          }
        }
      }
      if (typeof keyNode === "object" && _class(keyNode) === "[object Object]") {
        keyNode = "[object Object]";
      }
      keyNode = String(keyNode);
      if (_result === null) {
        _result = {};
      }
      if (keyTag === "tag:yaml.org,2002:merge") {
        if (Array.isArray(valueNode)) {
          if (valueNode.length > 100) {
            throwError(state, "abnormal merge sequence size");
          }
          for (let index = 0, quantity = valueNode.length; index < quantity; index += 1) {
            mergeMappings(state, _result, valueNode[index], overridableKeys);
          }
        } else {
          mergeMappings(state, _result, valueNode, overridableKeys);
        }
      } else {
        if (!state.json && !_hasOwnProperty2.call(overridableKeys, keyNode) && _hasOwnProperty2.call(_result, keyNode)) {
          state.line = startLine || state.line;
          state.lineStart = startLineStart || state.lineStart;
          state.position = startPos || state.position;
          throwError(state, "duplicated mapping key");
        }
        setProperty(_result, keyNode, valueNode);
        delete overridableKeys[keyNode];
      }
      return _result;
    }
    function readLineBreak(state) {
      const ch = state.input.charCodeAt(state.position);
      if (ch === 10) {
        state.position++;
      } else if (ch === 13) {
        state.position++;
        if (state.input.charCodeAt(state.position) === 10) {
          state.position++;
        }
      } else {
        throwError(state, "a line break is expected");
      }
      state.line += 1;
      state.lineStart = state.position;
      state.firstTabInLine = -1;
    }
    function skipSeparationSpace(state, allowComments, checkIndent) {
      let lineBreaks = 0;
      let ch = state.input.charCodeAt(state.position);
      while (ch !== 0) {
        while (isWhiteSpace(ch)) {
          if (ch === 9 && state.firstTabInLine === -1) {
            state.firstTabInLine = state.position;
          }
          ch = state.input.charCodeAt(++state.position);
        }
        if (allowComments && ch === 35) {
          do {
            ch = state.input.charCodeAt(++state.position);
          } while (ch !== 10 && ch !== 13 && ch !== 0);
        }
        if (isEol(ch)) {
          readLineBreak(state);
          ch = state.input.charCodeAt(state.position);
          lineBreaks++;
          state.lineIndent = 0;
          while (ch === 32) {
            state.lineIndent++;
            ch = state.input.charCodeAt(++state.position);
          }
        } else {
          break;
        }
      }
      if (checkIndent !== -1 && lineBreaks !== 0 && state.lineIndent < checkIndent) {
        throwWarning(state, "deficient indentation");
      }
      return lineBreaks;
    }
    function testDocumentSeparator(state) {
      let _position = state.position;
      let ch = state.input.charCodeAt(_position);
      if ((ch === 45 || ch === 46) && ch === state.input.charCodeAt(_position + 1) && ch === state.input.charCodeAt(_position + 2)) {
        _position += 3;
        ch = state.input.charCodeAt(_position);
        if (ch === 0 || isWsOrEol(ch)) {
          return true;
        }
      }
      return false;
    }
    function writeFoldedLines(state, count) {
      if (count === 1) {
        state.result += " ";
      } else if (count > 1) {
        state.result += common2.repeat("\n", count - 1);
      }
    }
    function readPlainScalar(state, nodeIndent, withinFlowCollection) {
      let captureStart;
      let captureEnd;
      let hasPendingContent;
      let _line;
      let _lineStart;
      let _lineIndent;
      const _kind = state.kind;
      const _result = state.result;
      let ch = state.input.charCodeAt(state.position);
      if (isWsOrEol(ch) || isFlowIndicator(ch) || ch === 35 || ch === 38 || ch === 42 || ch === 33 || ch === 124 || ch === 62 || ch === 39 || ch === 34 || ch === 37 || ch === 64 || ch === 96) {
        return false;
      }
      if (ch === 63 || ch === 45) {
        const following = state.input.charCodeAt(state.position + 1);
        if (isWsOrEol(following) || withinFlowCollection && isFlowIndicator(following)) {
          return false;
        }
      }
      state.kind = "scalar";
      state.result = "";
      captureStart = captureEnd = state.position;
      hasPendingContent = false;
      while (ch !== 0) {
        if (ch === 58) {
          const following = state.input.charCodeAt(state.position + 1);
          if (isWsOrEol(following) || withinFlowCollection && isFlowIndicator(following)) {
            break;
          }
        } else if (ch === 35) {
          const preceding = state.input.charCodeAt(state.position - 1);
          if (isWsOrEol(preceding)) {
            break;
          }
        } else if (state.position === state.lineStart && testDocumentSeparator(state) || withinFlowCollection && isFlowIndicator(ch)) {
          break;
        } else if (isEol(ch)) {
          _line = state.line;
          _lineStart = state.lineStart;
          _lineIndent = state.lineIndent;
          skipSeparationSpace(state, false, -1);
          if (state.lineIndent >= nodeIndent) {
            hasPendingContent = true;
            ch = state.input.charCodeAt(state.position);
            continue;
          } else {
            state.position = captureEnd;
            state.line = _line;
            state.lineStart = _lineStart;
            state.lineIndent = _lineIndent;
            break;
          }
        }
        if (hasPendingContent) {
          captureSegment(state, captureStart, captureEnd, false);
          writeFoldedLines(state, state.line - _line);
          captureStart = captureEnd = state.position;
          hasPendingContent = false;
        }
        if (!isWhiteSpace(ch)) {
          captureEnd = state.position + 1;
        }
        ch = state.input.charCodeAt(++state.position);
      }
      captureSegment(state, captureStart, captureEnd, false);
      if (state.result) {
        return true;
      }
      state.kind = _kind;
      state.result = _result;
      return false;
    }
    function readSingleQuotedScalar(state, nodeIndent) {
      let captureStart;
      let captureEnd;
      let ch = state.input.charCodeAt(state.position);
      if (ch !== 39) {
        return false;
      }
      state.kind = "scalar";
      state.result = "";
      state.position++;
      captureStart = captureEnd = state.position;
      while ((ch = state.input.charCodeAt(state.position)) !== 0) {
        if (ch === 39) {
          captureSegment(state, captureStart, state.position, true);
          ch = state.input.charCodeAt(++state.position);
          if (ch === 39) {
            captureStart = state.position;
            state.position++;
            captureEnd = state.position;
          } else {
            return true;
          }
        } else if (isEol(ch)) {
          captureSegment(state, captureStart, captureEnd, true);
          writeFoldedLines(state, skipSeparationSpace(state, false, nodeIndent));
          captureStart = captureEnd = state.position;
        } else if (state.position === state.lineStart && testDocumentSeparator(state)) {
          throwError(state, "unexpected end of the document within a single quoted scalar");
        } else {
          state.position++;
          if (!isWhiteSpace(ch)) {
            captureEnd = state.position;
          }
        }
      }
      throwError(state, "unexpected end of the stream within a single quoted scalar");
    }
    function readDoubleQuotedScalar(state, nodeIndent) {
      let captureStart;
      let captureEnd;
      let tmp;
      let ch = state.input.charCodeAt(state.position);
      if (ch !== 34) {
        return false;
      }
      state.kind = "scalar";
      state.result = "";
      state.position++;
      captureStart = captureEnd = state.position;
      while ((ch = state.input.charCodeAt(state.position)) !== 0) {
        if (ch === 34) {
          captureSegment(state, captureStart, state.position, true);
          state.position++;
          return true;
        } else if (ch === 92) {
          captureSegment(state, captureStart, state.position, true);
          ch = state.input.charCodeAt(++state.position);
          if (isEol(ch)) {
            skipSeparationSpace(state, false, nodeIndent);
          } else if (ch < 256 && simpleEscapeCheck[ch]) {
            state.result += simpleEscapeMap[ch];
            state.position++;
          } else if ((tmp = escapedHexLen(ch)) > 0) {
            let hexLength = tmp;
            let hexResult = 0;
            for (; hexLength > 0; hexLength--) {
              ch = state.input.charCodeAt(++state.position);
              if ((tmp = fromHexCode(ch)) >= 0) {
                hexResult = (hexResult << 4) + tmp;
              } else {
                throwError(state, "expected hexadecimal character");
              }
            }
            state.result += charFromCodepoint(hexResult);
            state.position++;
          } else {
            throwError(state, "unknown escape sequence");
          }
          captureStart = captureEnd = state.position;
        } else if (isEol(ch)) {
          captureSegment(state, captureStart, captureEnd, true);
          writeFoldedLines(state, skipSeparationSpace(state, false, nodeIndent));
          captureStart = captureEnd = state.position;
        } else if (state.position === state.lineStart && testDocumentSeparator(state)) {
          throwError(state, "unexpected end of the document within a double quoted scalar");
        } else {
          state.position++;
          if (!isWhiteSpace(ch)) {
            captureEnd = state.position;
          }
        }
      }
      throwError(state, "unexpected end of the stream within a double quoted scalar");
    }
    function readFlowCollection(state, nodeIndent) {
      let readNext = true;
      let _line;
      let _lineStart;
      let _pos;
      const _tag2 = state.tag;
      let _result;
      const _anchor = state.anchor;
      let terminator;
      let isPair;
      let isExplicitPair;
      let isMapping;
      const overridableKeys = /* @__PURE__ */ Object.create(null);
      let keyNode;
      let keyTag;
      let valueNode;
      let ch = state.input.charCodeAt(state.position);
      if (ch === 91) {
        terminator = 93;
        isMapping = false;
        _result = [];
      } else if (ch === 123) {
        terminator = 125;
        isMapping = true;
        _result = {};
      } else {
        return false;
      }
      if (state.anchor !== null) {
        storeAnchor(state, state.anchor, _result);
      }
      ch = state.input.charCodeAt(++state.position);
      while (ch !== 0) {
        skipSeparationSpace(state, true, nodeIndent);
        ch = state.input.charCodeAt(state.position);
        if (ch === terminator) {
          state.position++;
          state.tag = _tag2;
          state.anchor = _anchor;
          state.kind = isMapping ? "mapping" : "sequence";
          state.result = _result;
          return true;
        } else if (!readNext) {
          throwError(state, "missed comma between flow collection entries");
        } else if (ch === 44) {
          throwError(state, "expected the node content, but found ','");
        }
        keyTag = keyNode = valueNode = null;
        isPair = isExplicitPair = false;
        if (ch === 63) {
          const following = state.input.charCodeAt(state.position + 1);
          if (isWsOrEol(following)) {
            isPair = isExplicitPair = true;
            state.position++;
            skipSeparationSpace(state, true, nodeIndent);
          }
        }
        _line = state.line;
        _lineStart = state.lineStart;
        _pos = state.position;
        composeNode(state, nodeIndent, CONTEXT_FLOW_IN, false, true);
        keyTag = state.tag;
        keyNode = state.result;
        skipSeparationSpace(state, true, nodeIndent);
        ch = state.input.charCodeAt(state.position);
        if ((isExplicitPair || state.line === _line) && ch === 58) {
          isPair = true;
          ch = state.input.charCodeAt(++state.position);
          skipSeparationSpace(state, true, nodeIndent);
          composeNode(state, nodeIndent, CONTEXT_FLOW_IN, false, true);
          valueNode = state.result;
        }
        if (isMapping) {
          storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, valueNode, _line, _lineStart, _pos);
        } else if (isPair) {
          _result.push(storeMappingPair(state, null, overridableKeys, keyTag, keyNode, valueNode, _line, _lineStart, _pos));
        } else {
          _result.push(keyNode);
        }
        skipSeparationSpace(state, true, nodeIndent);
        ch = state.input.charCodeAt(state.position);
        if (ch === 44) {
          readNext = true;
          ch = state.input.charCodeAt(++state.position);
        } else {
          readNext = false;
        }
      }
      throwError(state, "unexpected end of the stream within a flow collection");
    }
    function readBlockScalar(state, nodeIndent) {
      let folding;
      let chomping = CHOMPING_CLIP;
      let didReadContent = false;
      let detectedIndent = false;
      let textIndent = nodeIndent;
      let emptyLines = 0;
      let atMoreIndented = false;
      let tmp;
      let ch = state.input.charCodeAt(state.position);
      if (ch === 124) {
        folding = false;
      } else if (ch === 62) {
        folding = true;
      } else {
        return false;
      }
      state.kind = "scalar";
      state.result = "";
      while (ch !== 0) {
        ch = state.input.charCodeAt(++state.position);
        if (ch === 43 || ch === 45) {
          if (CHOMPING_CLIP === chomping) {
            chomping = ch === 43 ? CHOMPING_KEEP : CHOMPING_STRIP;
          } else {
            throwError(state, "repeat of a chomping mode identifier");
          }
        } else if ((tmp = fromDecimalCode(ch)) >= 0) {
          if (tmp === 0) {
            throwError(state, "bad explicit indentation width of a block scalar; it cannot be less than one");
          } else if (!detectedIndent) {
            textIndent = nodeIndent + tmp - 1;
            detectedIndent = true;
          } else {
            throwError(state, "repeat of an indentation width identifier");
          }
        } else {
          break;
        }
      }
      if (isWhiteSpace(ch)) {
        do {
          ch = state.input.charCodeAt(++state.position);
        } while (isWhiteSpace(ch));
        if (ch === 35) {
          do {
            ch = state.input.charCodeAt(++state.position);
          } while (!isEol(ch) && ch !== 0);
        }
      }
      while (ch !== 0) {
        readLineBreak(state);
        state.lineIndent = 0;
        ch = state.input.charCodeAt(state.position);
        while ((!detectedIndent || state.lineIndent < textIndent) && ch === 32) {
          state.lineIndent++;
          ch = state.input.charCodeAt(++state.position);
        }
        if (!detectedIndent && state.lineIndent > textIndent) {
          textIndent = state.lineIndent;
        }
        if (isEol(ch)) {
          emptyLines++;
          continue;
        }
        if (!detectedIndent && textIndent === 0) {
          throwError(state, "missing indentation for block scalar");
        }
        if (state.lineIndent < textIndent) {
          if (chomping === CHOMPING_KEEP) {
            state.result += common2.repeat("\n", didReadContent ? 1 + emptyLines : emptyLines);
          } else if (chomping === CHOMPING_CLIP) {
            if (didReadContent) {
              state.result += "\n";
            }
          }
          break;
        }
        if (folding) {
          if (isWhiteSpace(ch)) {
            atMoreIndented = true;
            state.result += common2.repeat("\n", didReadContent ? 1 + emptyLines : emptyLines);
          } else if (atMoreIndented) {
            atMoreIndented = false;
            state.result += common2.repeat("\n", emptyLines + 1);
          } else if (emptyLines === 0) {
            if (didReadContent) {
              state.result += " ";
            }
          } else {
            state.result += common2.repeat("\n", emptyLines);
          }
        } else {
          state.result += common2.repeat("\n", didReadContent ? 1 + emptyLines : emptyLines);
        }
        didReadContent = true;
        detectedIndent = true;
        emptyLines = 0;
        const captureStart = state.position;
        while (!isEol(ch) && ch !== 0) {
          ch = state.input.charCodeAt(++state.position);
        }
        captureSegment(state, captureStart, state.position, false);
      }
      return true;
    }
    function readBlockSequence(state, nodeIndent) {
      const _tag2 = state.tag;
      const _anchor = state.anchor;
      const _result = [];
      let detected = false;
      if (state.firstTabInLine !== -1) return false;
      if (state.anchor !== null) {
        storeAnchor(state, state.anchor, _result);
      }
      let ch = state.input.charCodeAt(state.position);
      while (ch !== 0) {
        if (state.firstTabInLine !== -1) {
          state.position = state.firstTabInLine;
          throwError(state, "tab characters must not be used in indentation");
        }
        if (ch !== 45) {
          break;
        }
        const following = state.input.charCodeAt(state.position + 1);
        if (!isWsOrEol(following)) {
          break;
        }
        detected = true;
        state.position++;
        if (skipSeparationSpace(state, true, -1)) {
          if (state.lineIndent <= nodeIndent) {
            _result.push(null);
            ch = state.input.charCodeAt(state.position);
            continue;
          }
        }
        const _line = state.line;
        composeNode(state, nodeIndent, CONTEXT_BLOCK_IN, false, true);
        _result.push(state.result);
        skipSeparationSpace(state, true, -1);
        ch = state.input.charCodeAt(state.position);
        if ((state.line === _line || state.lineIndent > nodeIndent) && ch !== 0) {
          throwError(state, "bad indentation of a sequence entry");
        } else if (state.lineIndent < nodeIndent) {
          break;
        }
      }
      if (detected) {
        state.tag = _tag2;
        state.anchor = _anchor;
        state.kind = "sequence";
        state.result = _result;
        return true;
      }
      return false;
    }
    function readBlockMapping(state, nodeIndent, flowIndent) {
      let allowCompact;
      let _keyLine;
      let _keyLineStart;
      let _keyPos;
      const _tag2 = state.tag;
      const _anchor = state.anchor;
      const _result = {};
      const overridableKeys = /* @__PURE__ */ Object.create(null);
      let keyTag = null;
      let keyNode = null;
      let valueNode = null;
      let atExplicitKey = false;
      let detected = false;
      if (state.firstTabInLine !== -1) return false;
      if (state.anchor !== null) {
        storeAnchor(state, state.anchor, _result);
      }
      let ch = state.input.charCodeAt(state.position);
      while (ch !== 0) {
        if (!atExplicitKey && state.firstTabInLine !== -1) {
          state.position = state.firstTabInLine;
          throwError(state, "tab characters must not be used in indentation");
        }
        const following = state.input.charCodeAt(state.position + 1);
        const _line = state.line;
        if ((ch === 63 || ch === 58) && isWsOrEol(following)) {
          if (ch === 63) {
            if (atExplicitKey) {
              storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, null, _keyLine, _keyLineStart, _keyPos);
              keyTag = keyNode = valueNode = null;
            }
            detected = true;
            atExplicitKey = true;
            allowCompact = true;
          } else if (atExplicitKey) {
            atExplicitKey = false;
            allowCompact = true;
          } else {
            throwError(state, "incomplete explicit mapping pair; a key node is missed; or followed by a non-tabulated empty line");
          }
          state.position += 1;
          ch = following;
        } else {
          _keyLine = state.line;
          _keyLineStart = state.lineStart;
          _keyPos = state.position;
          if (!composeNode(state, flowIndent, CONTEXT_FLOW_OUT, false, true)) {
            break;
          }
          if (state.line === _line) {
            ch = state.input.charCodeAt(state.position);
            while (isWhiteSpace(ch)) {
              ch = state.input.charCodeAt(++state.position);
            }
            if (ch === 58) {
              ch = state.input.charCodeAt(++state.position);
              if (!isWsOrEol(ch)) {
                throwError(state, "a whitespace character is expected after the key-value separator within a block mapping");
              }
              if (atExplicitKey) {
                storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, null, _keyLine, _keyLineStart, _keyPos);
                keyTag = keyNode = valueNode = null;
              }
              detected = true;
              atExplicitKey = false;
              allowCompact = false;
              keyTag = state.tag;
              keyNode = state.result;
            } else if (detected) {
              throwError(state, "can not read an implicit mapping pair; a colon is missed");
            } else {
              state.tag = _tag2;
              state.anchor = _anchor;
              return true;
            }
          } else if (detected) {
            throwError(state, "can not read a block mapping entry; a multiline key may not be an implicit key");
          } else {
            state.tag = _tag2;
            state.anchor = _anchor;
            return true;
          }
        }
        if (state.line === _line || state.lineIndent > nodeIndent) {
          if (atExplicitKey) {
            _keyLine = state.line;
            _keyLineStart = state.lineStart;
            _keyPos = state.position;
          }
          if (composeNode(state, nodeIndent, CONTEXT_BLOCK_OUT, true, allowCompact)) {
            if (atExplicitKey) {
              keyNode = state.result;
            } else {
              valueNode = state.result;
            }
          }
          if (!atExplicitKey) {
            storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, valueNode, _keyLine, _keyLineStart, _keyPos);
            keyTag = keyNode = valueNode = null;
          }
          skipSeparationSpace(state, true, -1);
          ch = state.input.charCodeAt(state.position);
        }
        if ((state.line === _line || state.lineIndent > nodeIndent) && ch !== 0) {
          throwError(state, "bad indentation of a mapping entry");
        } else if (state.lineIndent < nodeIndent) {
          break;
        }
      }
      if (atExplicitKey) {
        storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, null, _keyLine, _keyLineStart, _keyPos);
      }
      if (detected) {
        state.tag = _tag2;
        state.anchor = _anchor;
        state.kind = "mapping";
        state.result = _result;
      }
      return detected;
    }
    function readTagProperty(state) {
      let isVerbatim = false;
      let isNamed = false;
      let tagHandle;
      let tagName;
      let ch = state.input.charCodeAt(state.position);
      if (ch !== 33) return false;
      if (state.tag !== null) {
        throwError(state, "duplication of a tag property");
      }
      ch = state.input.charCodeAt(++state.position);
      if (ch === 60) {
        isVerbatim = true;
        ch = state.input.charCodeAt(++state.position);
      } else if (ch === 33) {
        isNamed = true;
        tagHandle = "!!";
        ch = state.input.charCodeAt(++state.position);
      } else {
        tagHandle = "!";
      }
      let _position = state.position;
      if (isVerbatim) {
        do {
          ch = state.input.charCodeAt(++state.position);
        } while (ch !== 0 && ch !== 62);
        if (state.position < state.length) {
          tagName = state.input.slice(_position, state.position);
          ch = state.input.charCodeAt(++state.position);
        } else {
          throwError(state, "unexpected end of the stream within a verbatim tag");
        }
      } else {
        while (ch !== 0 && !isWsOrEol(ch)) {
          if (ch === 33) {
            if (!isNamed) {
              tagHandle = state.input.slice(_position - 1, state.position + 1);
              if (!PATTERN_TAG_HANDLE.test(tagHandle)) {
                throwError(state, "named tag handle cannot contain such characters");
              }
              isNamed = true;
              _position = state.position + 1;
            } else {
              throwError(state, "tag suffix cannot contain exclamation marks");
            }
          }
          ch = state.input.charCodeAt(++state.position);
        }
        tagName = state.input.slice(_position, state.position);
        if (PATTERN_FLOW_INDICATORS.test(tagName)) {
          throwError(state, "tag suffix cannot contain flow indicator characters");
        }
      }
      if (tagName && !PATTERN_TAG_URI.test(tagName)) {
        throwError(state, "tag name cannot contain such characters: " + tagName);
      }
      try {
        tagName = decodeURIComponent(tagName);
      } catch (err) {
        throwError(state, "tag name is malformed: " + tagName);
      }
      if (isVerbatim) {
        state.tag = tagName;
      } else if (_hasOwnProperty2.call(state.tagMap, tagHandle)) {
        state.tag = state.tagMap[tagHandle] + tagName;
      } else if (tagHandle === "!") {
        state.tag = "!" + tagName;
      } else if (tagHandle === "!!") {
        state.tag = "tag:yaml.org,2002:" + tagName;
      } else {
        throwError(state, 'undeclared tag handle "' + tagHandle + '"');
      }
      return true;
    }
    function readAnchorProperty(state) {
      let ch = state.input.charCodeAt(state.position);
      if (ch !== 38) return false;
      if (state.anchor !== null) {
        throwError(state, "duplication of an anchor property");
      }
      ch = state.input.charCodeAt(++state.position);
      const _position = state.position;
      while (ch !== 0 && !isWsOrEol(ch) && !isFlowIndicator(ch)) {
        ch = state.input.charCodeAt(++state.position);
      }
      if (state.position === _position) {
        throwError(state, "name of an anchor node must contain at least one character");
      }
      state.anchor = state.input.slice(_position, state.position);
      return true;
    }
    function readAlias(state) {
      let ch = state.input.charCodeAt(state.position);
      if (ch !== 42) return false;
      ch = state.input.charCodeAt(++state.position);
      const _position = state.position;
      while (ch !== 0 && !isWsOrEol(ch) && !isFlowIndicator(ch)) {
        ch = state.input.charCodeAt(++state.position);
      }
      if (state.position === _position) {
        throwError(state, "name of an alias node must contain at least one character");
      }
      const alias = state.input.slice(_position, state.position);
      if (!_hasOwnProperty2.call(state.anchorMap, alias)) {
        throwError(state, 'unidentified alias "' + alias + '"');
      }
      state.result = state.anchorMap[alias];
      skipSeparationSpace(state, true, -1);
      return true;
    }
    function tryReadBlockMappingFromProperty(state, propertyStart, nodeIndent, flowIndent) {
      const fallbackState = snapshotState(state);
      beginAnchorTransaction(state);
      restoreState(state, propertyStart);
      state.tag = null;
      state.anchor = null;
      state.kind = null;
      state.result = null;
      if (readBlockMapping(state, nodeIndent, flowIndent) && state.kind === "mapping") {
        commitAnchorTransaction(state);
        return true;
      }
      rollbackAnchorTransaction(state);
      restoreState(state, fallbackState);
      return false;
    }
    function composeNode(state, parentIndent, nodeContext, allowToSeek, allowCompact) {
      let allowBlockScalars;
      let allowBlockCollections;
      let indentStatus = 1;
      let atNewLine = false;
      let hasContent = false;
      let propertyStart = null;
      let type2;
      let flowIndent;
      let blockIndent;
      if (state.depth >= state.maxDepth) {
        throwError(state, "nesting exceeded maxDepth (" + state.maxDepth + ")");
      }
      state.depth += 1;
      if (state.listener !== null) {
        state.listener("open", state);
      }
      state.tag = null;
      state.anchor = null;
      state.kind = null;
      state.result = null;
      const allowBlockStyles = allowBlockScalars = allowBlockCollections = CONTEXT_BLOCK_OUT === nodeContext || CONTEXT_BLOCK_IN === nodeContext;
      if (allowToSeek) {
        if (skipSeparationSpace(state, true, -1)) {
          atNewLine = true;
          if (state.lineIndent > parentIndent) {
            indentStatus = 1;
          } else if (state.lineIndent === parentIndent) {
            indentStatus = 0;
          } else if (state.lineIndent < parentIndent) {
            indentStatus = -1;
          }
        }
      }
      if (indentStatus === 1) {
        while (true) {
          const ch = state.input.charCodeAt(state.position);
          const propertyState = snapshotState(state);
          if (atNewLine && (ch === 33 && state.tag !== null || ch === 38 && state.anchor !== null)) {
            break;
          }
          if (!readTagProperty(state) && !readAnchorProperty(state)) {
            break;
          }
          if (propertyStart === null) {
            propertyStart = propertyState;
          }
          if (skipSeparationSpace(state, true, -1)) {
            atNewLine = true;
            allowBlockCollections = allowBlockStyles;
            if (state.lineIndent > parentIndent) {
              indentStatus = 1;
            } else if (state.lineIndent === parentIndent) {
              indentStatus = 0;
            } else if (state.lineIndent < parentIndent) {
              indentStatus = -1;
            }
          } else {
            allowBlockCollections = false;
          }
        }
      }
      if (allowBlockCollections) {
        allowBlockCollections = atNewLine || allowCompact;
      }
      if (indentStatus === 1 || CONTEXT_BLOCK_OUT === nodeContext) {
        if (CONTEXT_FLOW_IN === nodeContext || CONTEXT_FLOW_OUT === nodeContext) {
          flowIndent = parentIndent;
        } else {
          flowIndent = parentIndent + 1;
        }
        blockIndent = state.position - state.lineStart;
        if (indentStatus === 1) {
          if (allowBlockCollections && (readBlockSequence(state, blockIndent) || readBlockMapping(state, blockIndent, flowIndent)) || readFlowCollection(state, flowIndent)) {
            hasContent = true;
          } else {
            const ch = state.input.charCodeAt(state.position);
            if (propertyStart !== null && allowBlockStyles && !allowBlockCollections && ch !== 124 && ch !== 62 && tryReadBlockMappingFromProperty(
              state,
              propertyStart,
              propertyStart.position - propertyStart.lineStart,
              flowIndent
            )) {
              hasContent = true;
            } else if (allowBlockScalars && readBlockScalar(state, flowIndent) || readSingleQuotedScalar(state, flowIndent) || readDoubleQuotedScalar(state, flowIndent)) {
              hasContent = true;
            } else if (readAlias(state)) {
              hasContent = true;
              if (state.tag !== null || state.anchor !== null) {
                throwError(state, "alias node should not have any properties");
              }
            } else if (readPlainScalar(state, flowIndent, CONTEXT_FLOW_IN === nodeContext)) {
              hasContent = true;
              if (state.tag === null) {
                state.tag = "?";
              }
            }
            if (state.anchor !== null) {
              storeAnchor(state, state.anchor, state.result);
            }
          }
        } else if (indentStatus === 0) {
          hasContent = allowBlockCollections && readBlockSequence(state, blockIndent);
        }
      }
      if (state.tag === null) {
        if (state.anchor !== null) {
          storeAnchor(state, state.anchor, state.result);
        }
      } else if (state.tag === "?") {
        if (state.result !== null && state.kind !== "scalar") {
          throwError(state, 'unacceptable node kind for !<?> tag; it should be "scalar", not "' + state.kind + '"');
        }
        for (let typeIndex = 0, typeQuantity = state.implicitTypes.length; typeIndex < typeQuantity; typeIndex += 1) {
          type2 = state.implicitTypes[typeIndex];
          if (type2.resolve(state.result)) {
            state.result = type2.construct(state.result);
            state.tag = type2.tag;
            if (state.anchor !== null) {
              storeAnchor(state, state.anchor, state.result);
            }
            break;
          }
        }
      } else if (state.tag !== "!") {
        if (_hasOwnProperty2.call(state.typeMap[state.kind || "fallback"], state.tag)) {
          type2 = state.typeMap[state.kind || "fallback"][state.tag];
        } else {
          type2 = null;
          const typeList = state.typeMap.multi[state.kind || "fallback"];
          for (let typeIndex = 0, typeQuantity = typeList.length; typeIndex < typeQuantity; typeIndex += 1) {
            if (state.tag.slice(0, typeList[typeIndex].tag.length) === typeList[typeIndex].tag) {
              type2 = typeList[typeIndex];
              break;
            }
          }
        }
        if (!type2) {
          throwError(state, "unknown tag !<" + state.tag + ">");
        }
        if (state.result !== null && type2.kind !== state.kind) {
          throwError(state, "unacceptable node kind for !<" + state.tag + '> tag; it should be "' + type2.kind + '", not "' + state.kind + '"');
        }
        if (!type2.resolve(state.result, state.tag)) {
          throwError(state, "cannot resolve a node with !<" + state.tag + "> explicit tag");
        } else {
          state.result = type2.construct(state.result, state.tag);
          if (state.anchor !== null) {
            storeAnchor(state, state.anchor, state.result);
          }
        }
      }
      if (state.listener !== null) {
        state.listener("close", state);
      }
      state.depth -= 1;
      return state.tag !== null || state.anchor !== null || hasContent;
    }
    function readDocument(state) {
      const documentStart = state.position;
      let hasDirectives = false;
      let ch;
      state.version = null;
      state.checkLineBreaks = state.legacy;
      state.tagMap = /* @__PURE__ */ Object.create(null);
      state.anchorMap = /* @__PURE__ */ Object.create(null);
      while ((ch = state.input.charCodeAt(state.position)) !== 0) {
        skipSeparationSpace(state, true, -1);
        ch = state.input.charCodeAt(state.position);
        if (state.lineIndent > 0 || ch !== 37) {
          break;
        }
        hasDirectives = true;
        ch = state.input.charCodeAt(++state.position);
        let _position = state.position;
        while (ch !== 0 && !isWsOrEol(ch)) {
          ch = state.input.charCodeAt(++state.position);
        }
        const directiveName = state.input.slice(_position, state.position);
        const directiveArgs = [];
        if (directiveName.length < 1) {
          throwError(state, "directive name must not be less than one character in length");
        }
        while (ch !== 0) {
          while (isWhiteSpace(ch)) {
            ch = state.input.charCodeAt(++state.position);
          }
          if (ch === 35) {
            do {
              ch = state.input.charCodeAt(++state.position);
            } while (ch !== 0 && !isEol(ch));
            break;
          }
          if (isEol(ch)) break;
          _position = state.position;
          while (ch !== 0 && !isWsOrEol(ch)) {
            ch = state.input.charCodeAt(++state.position);
          }
          directiveArgs.push(state.input.slice(_position, state.position));
        }
        if (ch !== 0) readLineBreak(state);
        if (_hasOwnProperty2.call(directiveHandlers, directiveName)) {
          directiveHandlers[directiveName](state, directiveName, directiveArgs);
        } else {
          throwWarning(state, 'unknown document directive "' + directiveName + '"');
        }
      }
      skipSeparationSpace(state, true, -1);
      if (state.lineIndent === 0 && state.input.charCodeAt(state.position) === 45 && state.input.charCodeAt(state.position + 1) === 45 && state.input.charCodeAt(state.position + 2) === 45) {
        state.position += 3;
        skipSeparationSpace(state, true, -1);
      } else if (hasDirectives) {
        throwError(state, "directives end mark is expected");
      }
      composeNode(state, state.lineIndent - 1, CONTEXT_BLOCK_OUT, false, true);
      skipSeparationSpace(state, true, -1);
      if (state.checkLineBreaks && PATTERN_NON_ASCII_LINE_BREAKS.test(state.input.slice(documentStart, state.position))) {
        throwWarning(state, "non-ASCII line breaks are interpreted as content");
      }
      state.documents.push(state.result);
      if (state.position === state.lineStart && testDocumentSeparator(state)) {
        if (state.input.charCodeAt(state.position) === 46) {
          state.position += 3;
          skipSeparationSpace(state, true, -1);
        }
        return;
      }
      if (state.position < state.length - 1) {
        throwError(state, "end of the stream or a document separator is expected");
      }
    }
    function loadDocuments(input, options2) {
      input = String(input);
      options2 = options2 || {};
      if (input.length !== 0) {
        if (input.charCodeAt(input.length - 1) !== 10 && input.charCodeAt(input.length - 1) !== 13) {
          input += "\n";
        }
        if (input.charCodeAt(0) === 65279) {
          input = input.slice(1);
        }
      }
      const state = new State(input, options2);
      const nullpos = input.indexOf("\0");
      if (nullpos !== -1) {
        state.position = nullpos;
        throwError(state, "null byte is not allowed in input");
      }
      state.input += "\0";
      while (state.input.charCodeAt(state.position) === 32) {
        state.lineIndent += 1;
        state.position += 1;
      }
      while (state.position < state.length - 1) {
        readDocument(state);
      }
      return state.documents;
    }
    function loadAll2(input, iterator, options2) {
      if (iterator !== null && typeof iterator === "object" && typeof options2 === "undefined") {
        options2 = iterator;
        iterator = null;
      }
      const documents = loadDocuments(input, options2);
      if (typeof iterator !== "function") {
        return documents;
      }
      for (let index = 0, length = documents.length; index < length; index += 1) {
        iterator(documents[index]);
      }
    }
    function load2(input, options2) {
      const documents = loadDocuments(input, options2);
      if (documents.length === 0) {
        return void 0;
      } else if (documents.length === 1) {
        return documents[0];
      }
      throw new YAMLException2("expected a single document in the stream, but found more");
    }
    loader.loadAll = loadAll2;
    loader.load = load2;
    return loader;
  }
  var dumper = {};
  var hasRequiredDumper;
  function requireDumper() {
    if (hasRequiredDumper) return dumper;
    hasRequiredDumper = 1;
    const common2 = requireCommon();
    const YAMLException2 = requireException();
    const DEFAULT_SCHEMA2 = require_default();
    const _toString = Object.prototype.toString;
    const _hasOwnProperty2 = Object.prototype.hasOwnProperty;
    const CHAR_BOM = 65279;
    const CHAR_TAB = 9;
    const CHAR_LINE_FEED = 10;
    const CHAR_CARRIAGE_RETURN = 13;
    const CHAR_SPACE = 32;
    const CHAR_EXCLAMATION = 33;
    const CHAR_DOUBLE_QUOTE = 34;
    const CHAR_SHARP = 35;
    const CHAR_PERCENT = 37;
    const CHAR_AMPERSAND = 38;
    const CHAR_SINGLE_QUOTE = 39;
    const CHAR_ASTERISK = 42;
    const CHAR_COMMA = 44;
    const CHAR_MINUS = 45;
    const CHAR_COLON = 58;
    const CHAR_EQUALS = 61;
    const CHAR_GREATER_THAN = 62;
    const CHAR_QUESTION = 63;
    const CHAR_COMMERCIAL_AT = 64;
    const CHAR_LEFT_SQUARE_BRACKET = 91;
    const CHAR_RIGHT_SQUARE_BRACKET = 93;
    const CHAR_GRAVE_ACCENT = 96;
    const CHAR_LEFT_CURLY_BRACKET = 123;
    const CHAR_VERTICAL_LINE = 124;
    const CHAR_RIGHT_CURLY_BRACKET = 125;
    const ESCAPE_SEQUENCES = {};
    ESCAPE_SEQUENCES[0] = "\\0";
    ESCAPE_SEQUENCES[7] = "\\a";
    ESCAPE_SEQUENCES[8] = "\\b";
    ESCAPE_SEQUENCES[9] = "\\t";
    ESCAPE_SEQUENCES[10] = "\\n";
    ESCAPE_SEQUENCES[11] = "\\v";
    ESCAPE_SEQUENCES[12] = "\\f";
    ESCAPE_SEQUENCES[13] = "\\r";
    ESCAPE_SEQUENCES[27] = "\\e";
    ESCAPE_SEQUENCES[34] = '\\"';
    ESCAPE_SEQUENCES[92] = "\\\\";
    ESCAPE_SEQUENCES[133] = "\\N";
    ESCAPE_SEQUENCES[160] = "\\_";
    ESCAPE_SEQUENCES[8232] = "\\L";
    ESCAPE_SEQUENCES[8233] = "\\P";
    const DEPRECATED_BOOLEANS_SYNTAX = [
      "y",
      "Y",
      "yes",
      "Yes",
      "YES",
      "on",
      "On",
      "ON",
      "n",
      "N",
      "no",
      "No",
      "NO",
      "off",
      "Off",
      "OFF"
    ];
    const DEPRECATED_BASE60_SYNTAX = /^[-+]?[0-9_]+(?::[0-9_]+)+(?:\.[0-9_]*)?$/;
    function compileStyleMap(schema2, map2) {
      if (map2 === null) return {};
      const result = {};
      const keys = Object.keys(map2);
      for (let index = 0, length = keys.length; index < length; index += 1) {
        let tag2 = keys[index];
        let style = String(map2[tag2]);
        if (tag2.slice(0, 2) === "!!") {
          tag2 = "tag:yaml.org,2002:" + tag2.slice(2);
        }
        const type2 = schema2.compiledTypeMap["fallback"][tag2];
        if (type2 && _hasOwnProperty2.call(type2.styleAliases, style)) {
          style = type2.styleAliases[style];
        }
        result[tag2] = style;
      }
      return result;
    }
    function encodeHex(character) {
      let handle;
      let length;
      const string = character.toString(16).toUpperCase();
      if (character <= 255) {
        handle = "x";
        length = 2;
      } else if (character <= 65535) {
        handle = "u";
        length = 4;
      } else if (character <= 4294967295) {
        handle = "U";
        length = 8;
      } else {
        throw new YAMLException2("code point within a string may not be greater than 0xFFFFFFFF");
      }
      return "\\" + handle + common2.repeat("0", length - string.length) + string;
    }
    const QUOTING_TYPE_SINGLE = 1;
    const QUOTING_TYPE_DOUBLE = 2;
    function State(options2) {
      this.schema = options2["schema"] || DEFAULT_SCHEMA2;
      this.indent = Math.max(1, options2["indent"] || 2);
      this.noArrayIndent = options2["noArrayIndent"] || false;
      this.skipInvalid = options2["skipInvalid"] || false;
      this.flowLevel = common2.isNothing(options2["flowLevel"]) ? -1 : options2["flowLevel"];
      this.styleMap = compileStyleMap(this.schema, options2["styles"] || null);
      this.sortKeys = options2["sortKeys"] || false;
      this.lineWidth = options2["lineWidth"] || 80;
      this.noRefs = options2["noRefs"] || false;
      this.noCompatMode = options2["noCompatMode"] || false;
      this.condenseFlow = options2["condenseFlow"] || false;
      this.quotingType = options2["quotingType"] === '"' ? QUOTING_TYPE_DOUBLE : QUOTING_TYPE_SINGLE;
      this.forceQuotes = options2["forceQuotes"] || false;
      this.replacer = typeof options2["replacer"] === "function" ? options2["replacer"] : null;
      this.implicitTypes = this.schema.compiledImplicit;
      this.explicitTypes = this.schema.compiledExplicit;
      this.tag = null;
      this.result = "";
      this.duplicates = [];
      this.usedDuplicates = null;
    }
    function indentString(string, spaces) {
      const ind = common2.repeat(" ", spaces);
      let position = 0;
      let result = "";
      const length = string.length;
      while (position < length) {
        let line;
        const next = string.indexOf("\n", position);
        if (next === -1) {
          line = string.slice(position);
          position = length;
        } else {
          line = string.slice(position, next + 1);
          position = next + 1;
        }
        if (line.length && line !== "\n") result += ind;
        result += line;
      }
      return result;
    }
    function generateNextLine(state, level) {
      return "\n" + common2.repeat(" ", state.indent * level);
    }
    function testImplicitResolving(state, str2) {
      for (let index = 0, length = state.implicitTypes.length; index < length; index += 1) {
        const type2 = state.implicitTypes[index];
        if (type2.resolve(str2)) {
          return true;
        }
      }
      return false;
    }
    function isWhitespace(c) {
      return c === CHAR_SPACE || c === CHAR_TAB;
    }
    function isPrintable(c) {
      return c >= 32 && c <= 126 || c >= 161 && c <= 55295 && c !== 8232 && c !== 8233 || c >= 57344 && c <= 65533 && c !== CHAR_BOM || c >= 65536 && c <= 1114111;
    }
    function isNsCharOrWhitespace(c) {
      return isPrintable(c) && c !== CHAR_BOM && // - b-char
      c !== CHAR_CARRIAGE_RETURN && c !== CHAR_LINE_FEED;
    }
    function isPlainSafe(c, prev, inblock) {
      const cIsNsCharOrWhitespace = isNsCharOrWhitespace(c);
      const cIsNsChar = cIsNsCharOrWhitespace && !isWhitespace(c);
      return (
        // ns-plain-safe
        (inblock ? cIsNsCharOrWhitespace : cIsNsCharOrWhitespace && // - c-flow-indicator
        c !== CHAR_COMMA && c !== CHAR_LEFT_SQUARE_BRACKET && c !== CHAR_RIGHT_SQUARE_BRACKET && c !== CHAR_LEFT_CURLY_BRACKET && c !== CHAR_RIGHT_CURLY_BRACKET) && // ns-plain-char
        c !== CHAR_SHARP && // false on '#'
        !(prev === CHAR_COLON && !cIsNsChar) || // false on ': '
        isNsCharOrWhitespace(prev) && !isWhitespace(prev) && c === CHAR_SHARP || // change to true on '[^ ]#'
        prev === CHAR_COLON && cIsNsChar
      );
    }
    function isPlainSafeFirst(c) {
      return isPrintable(c) && c !== CHAR_BOM && !isWhitespace(c) && // - s-white
      // - (c-indicator ::=
      // “-” | “?” | “:” | “,” | “[” | “]” | “{” | “}”
      c !== CHAR_MINUS && c !== CHAR_QUESTION && c !== CHAR_COLON && c !== CHAR_COMMA && c !== CHAR_LEFT_SQUARE_BRACKET && c !== CHAR_RIGHT_SQUARE_BRACKET && c !== CHAR_LEFT_CURLY_BRACKET && c !== CHAR_RIGHT_CURLY_BRACKET && // | “#” | “&” | “*” | “!” | “|” | “=” | “>” | “'” | “"”
      c !== CHAR_SHARP && c !== CHAR_AMPERSAND && c !== CHAR_ASTERISK && c !== CHAR_EXCLAMATION && c !== CHAR_VERTICAL_LINE && c !== CHAR_EQUALS && c !== CHAR_GREATER_THAN && c !== CHAR_SINGLE_QUOTE && c !== CHAR_DOUBLE_QUOTE && // | “%” | “@” | “`”)
      c !== CHAR_PERCENT && c !== CHAR_COMMERCIAL_AT && c !== CHAR_GRAVE_ACCENT;
    }
    function isPlainSafeLast(c) {
      return !isWhitespace(c) && c !== CHAR_COLON;
    }
    function codePointAt(string, pos) {
      const first = string.charCodeAt(pos);
      let second;
      if (first >= 55296 && first <= 56319 && pos + 1 < string.length) {
        second = string.charCodeAt(pos + 1);
        if (second >= 56320 && second <= 57343) {
          return (first - 55296) * 1024 + second - 56320 + 65536;
        }
      }
      return first;
    }
    function needIndentIndicator(string) {
      const leadingSpaceRe = /^\n* /;
      return leadingSpaceRe.test(string);
    }
    const STYLE_PLAIN = 1;
    const STYLE_SINGLE = 2;
    const STYLE_LITERAL = 3;
    const STYLE_FOLDED = 4;
    const STYLE_DOUBLE = 5;
    function chooseScalarStyle(string, singleLineOnly, indentPerLevel, lineWidth, testAmbiguousType, quotingType, forceQuotes, inblock) {
      let i;
      let char = 0;
      let prevChar = null;
      let hasLineBreak = false;
      let hasFoldableLine = false;
      const shouldTrackWidth = lineWidth !== -1;
      let previousLineBreak = -1;
      let plain = isPlainSafeFirst(codePointAt(string, 0)) && isPlainSafeLast(codePointAt(string, string.length - 1));
      if (singleLineOnly || forceQuotes) {
        for (i = 0; i < string.length; char >= 65536 ? i += 2 : i++) {
          char = codePointAt(string, i);
          if (!isPrintable(char)) {
            return STYLE_DOUBLE;
          }
          plain = plain && isPlainSafe(char, prevChar, inblock);
          prevChar = char;
        }
      } else {
        for (i = 0; i < string.length; char >= 65536 ? i += 2 : i++) {
          char = codePointAt(string, i);
          if (char === CHAR_LINE_FEED) {
            hasLineBreak = true;
            if (shouldTrackWidth) {
              hasFoldableLine = hasFoldableLine || // Foldable line = too long, and not more-indented.
              i - previousLineBreak - 1 > lineWidth && string[previousLineBreak + 1] !== " ";
              previousLineBreak = i;
            }
          } else if (!isPrintable(char)) {
            return STYLE_DOUBLE;
          }
          plain = plain && isPlainSafe(char, prevChar, inblock);
          prevChar = char;
        }
        hasFoldableLine = hasFoldableLine || shouldTrackWidth && (i - previousLineBreak - 1 > lineWidth && string[previousLineBreak + 1] !== " ");
      }
      if (!hasLineBreak && !hasFoldableLine) {
        if (plain && !forceQuotes && !testAmbiguousType(string)) {
          return STYLE_PLAIN;
        }
        return quotingType === QUOTING_TYPE_DOUBLE ? STYLE_DOUBLE : STYLE_SINGLE;
      }
      if (indentPerLevel > 9 && needIndentIndicator(string)) {
        return STYLE_DOUBLE;
      }
      if (!forceQuotes) {
        return hasFoldableLine ? STYLE_FOLDED : STYLE_LITERAL;
      }
      return quotingType === QUOTING_TYPE_DOUBLE ? STYLE_DOUBLE : STYLE_SINGLE;
    }
    function writeScalar(state, string, level, iskey, inblock) {
      state.dump = (function() {
        if (string.length === 0) {
          return state.quotingType === QUOTING_TYPE_DOUBLE ? '""' : "''";
        }
        if (!state.noCompatMode) {
          if (DEPRECATED_BOOLEANS_SYNTAX.indexOf(string) !== -1 || DEPRECATED_BASE60_SYNTAX.test(string)) {
            return state.quotingType === QUOTING_TYPE_DOUBLE ? '"' + string + '"' : "'" + string + "'";
          }
        }
        const indent = state.indent * Math.max(1, level);
        const lineWidth = state.lineWidth === -1 ? -1 : Math.max(Math.min(state.lineWidth, 40), state.lineWidth - indent);
        const singleLineOnly = iskey || // No block styles in flow mode.
        state.flowLevel > -1 && level >= state.flowLevel;
        function testAmbiguity(string2) {
          return testImplicitResolving(state, string2);
        }
        switch (chooseScalarStyle(
          string,
          singleLineOnly,
          state.indent,
          lineWidth,
          testAmbiguity,
          state.quotingType,
          state.forceQuotes && !iskey,
          inblock
        )) {
          case STYLE_PLAIN:
            return string;
          case STYLE_SINGLE:
            return "'" + string.replace(/'/g, "''") + "'";
          case STYLE_LITERAL:
            return "|" + blockHeader(string, state.indent) + dropEndingNewline(indentString(string, indent));
          case STYLE_FOLDED:
            return ">" + blockHeader(string, state.indent) + dropEndingNewline(indentString(foldString(string, lineWidth), indent));
          case STYLE_DOUBLE:
            return '"' + escapeString(string) + '"';
          default:
            throw new YAMLException2("impossible error: invalid scalar style");
        }
      })();
    }
    function blockHeader(string, indentPerLevel) {
      const indentIndicator = needIndentIndicator(string) ? String(indentPerLevel) : "";
      const clip = string[string.length - 1] === "\n";
      const keep = clip && (string[string.length - 2] === "\n" || string === "\n");
      const chomp = keep ? "+" : clip ? "" : "-";
      return indentIndicator + chomp + "\n";
    }
    function dropEndingNewline(string) {
      return string[string.length - 1] === "\n" ? string.slice(0, -1) : string;
    }
    function foldString(string, width) {
      const lineRe = /(\n+)([^\n]*)/g;
      let result = (function() {
        let nextLF = string.indexOf("\n");
        nextLF = nextLF !== -1 ? nextLF : string.length;
        lineRe.lastIndex = nextLF;
        return foldLine(string.slice(0, nextLF), width);
      })();
      let prevMoreIndented = string[0] === "\n" || string[0] === " ";
      let moreIndented;
      let match;
      while (match = lineRe.exec(string)) {
        const prefix = match[1];
        const line = match[2];
        moreIndented = line[0] === " ";
        result += prefix + (!prevMoreIndented && !moreIndented && line !== "" ? "\n" : "") + foldLine(line, width);
        prevMoreIndented = moreIndented;
      }
      return result;
    }
    function foldLine(line, width) {
      if (line === "" || line[0] === " ") return line;
      const breakRe = / [^ ]/g;
      let match;
      let start = 0;
      let end;
      let curr = 0;
      let next = 0;
      let result = "";
      while (match = breakRe.exec(line)) {
        next = match.index;
        if (next - start > width) {
          end = curr > start ? curr : next;
          result += "\n" + line.slice(start, end);
          start = end + 1;
        }
        curr = next;
      }
      result += "\n";
      if (line.length - start > width && curr > start) {
        result += line.slice(start, curr) + "\n" + line.slice(curr + 1);
      } else {
        result += line.slice(start);
      }
      return result.slice(1);
    }
    function escapeString(string) {
      let result = "";
      let char = 0;
      for (let i = 0; i < string.length; char >= 65536 ? i += 2 : i++) {
        char = codePointAt(string, i);
        const escapeSeq = ESCAPE_SEQUENCES[char];
        if (!escapeSeq && isPrintable(char)) {
          result += string[i];
          if (char >= 65536) result += string[i + 1];
        } else {
          result += escapeSeq || encodeHex(char);
        }
      }
      return result;
    }
    function writeFlowSequence(state, level, object) {
      let _result = "";
      const _tag2 = state.tag;
      for (let index = 0, length = object.length; index < length; index += 1) {
        let value = object[index];
        if (state.replacer) {
          value = state.replacer.call(object, String(index), value);
        }
        if (writeNode(state, level, value, false, false) || typeof value === "undefined" && writeNode(state, level, null, false, false)) {
          if (_result !== "") _result += "," + (!state.condenseFlow ? " " : "");
          _result += state.dump;
        }
      }
      state.tag = _tag2;
      state.dump = "[" + _result + "]";
    }
    function writeBlockSequence(state, level, object, compact) {
      let _result = "";
      const _tag2 = state.tag;
      for (let index = 0, length = object.length; index < length; index += 1) {
        let value = object[index];
        if (state.replacer) {
          value = state.replacer.call(object, String(index), value);
        }
        if (writeNode(state, level + 1, value, true, true, false, true) || typeof value === "undefined" && writeNode(state, level + 1, null, true, true, false, true)) {
          if (!compact || _result !== "") {
            _result += generateNextLine(state, level);
          }
          if (state.dump && CHAR_LINE_FEED === state.dump.charCodeAt(0)) {
            _result += "-";
          } else {
            _result += "- ";
          }
          _result += state.dump;
        }
      }
      state.tag = _tag2;
      state.dump = _result || "[]";
    }
    function writeFlowMapping(state, level, object) {
      let _result = "";
      const _tag2 = state.tag;
      const objectKeyList = Object.keys(object);
      for (let index = 0, length = objectKeyList.length; index < length; index += 1) {
        let pairBuffer = "";
        if (_result !== "") pairBuffer += ", ";
        if (state.condenseFlow) pairBuffer += '"';
        const objectKey = objectKeyList[index];
        let objectValue = object[objectKey];
        if (state.replacer) {
          objectValue = state.replacer.call(object, objectKey, objectValue);
        }
        if (!writeNode(state, level, objectKey, false, false)) {
          continue;
        }
        if (state.dump.length > 1024) pairBuffer += "? ";
        pairBuffer += state.dump + (state.condenseFlow ? '"' : "") + ":" + (state.condenseFlow ? "" : " ");
        if (!writeNode(state, level, objectValue, false, false)) {
          continue;
        }
        pairBuffer += state.dump;
        _result += pairBuffer;
      }
      state.tag = _tag2;
      state.dump = "{" + _result + "}";
    }
    function writeBlockMapping(state, level, object, compact) {
      let _result = "";
      const _tag2 = state.tag;
      const objectKeyList = Object.keys(object);
      if (state.sortKeys === true) {
        objectKeyList.sort();
      } else if (typeof state.sortKeys === "function") {
        objectKeyList.sort(state.sortKeys);
      } else if (state.sortKeys) {
        throw new YAMLException2("sortKeys must be a boolean or a function");
      }
      for (let index = 0, length = objectKeyList.length; index < length; index += 1) {
        let pairBuffer = "";
        if (!compact || _result !== "") {
          pairBuffer += generateNextLine(state, level);
        }
        const objectKey = objectKeyList[index];
        let objectValue = object[objectKey];
        if (state.replacer) {
          objectValue = state.replacer.call(object, objectKey, objectValue);
        }
        if (!writeNode(state, level + 1, objectKey, true, true, true)) {
          continue;
        }
        const explicitPair = state.tag !== null && state.tag !== "?" || state.dump && state.dump.length > 1024;
        if (explicitPair) {
          if (state.dump && CHAR_LINE_FEED === state.dump.charCodeAt(0)) {
            pairBuffer += "?";
          } else {
            pairBuffer += "? ";
          }
        }
        pairBuffer += state.dump;
        if (explicitPair) {
          pairBuffer += generateNextLine(state, level);
        }
        if (!writeNode(state, level + 1, objectValue, true, explicitPair)) {
          continue;
        }
        if (state.dump && CHAR_LINE_FEED === state.dump.charCodeAt(0)) {
          pairBuffer += ":";
        } else {
          pairBuffer += ": ";
        }
        pairBuffer += state.dump;
        _result += pairBuffer;
      }
      state.tag = _tag2;
      state.dump = _result || "{}";
    }
    function detectType(state, object, explicit) {
      const typeList = explicit ? state.explicitTypes : state.implicitTypes;
      for (let index = 0, length = typeList.length; index < length; index += 1) {
        const type2 = typeList[index];
        if ((type2.instanceOf || type2.predicate) && (!type2.instanceOf || typeof object === "object" && object instanceof type2.instanceOf) && (!type2.predicate || type2.predicate(object))) {
          if (explicit) {
            if (type2.multi && type2.representName) {
              state.tag = type2.representName(object);
            } else {
              state.tag = type2.tag;
            }
          } else {
            state.tag = "?";
          }
          if (type2.represent) {
            const style = state.styleMap[type2.tag] || type2.defaultStyle;
            let _result;
            if (_toString.call(type2.represent) === "[object Function]") {
              _result = type2.represent(object, style);
            } else if (_hasOwnProperty2.call(type2.represent, style)) {
              _result = type2.represent[style](object, style);
            } else {
              throw new YAMLException2("!<" + type2.tag + '> tag resolver accepts not "' + style + '" style');
            }
            state.dump = _result;
          }
          return true;
        }
      }
      return false;
    }
    function writeNode(state, level, object, block2, compact, iskey, isblockseq) {
      state.tag = null;
      state.dump = object;
      if (!detectType(state, object, false)) {
        detectType(state, object, true);
      }
      const type2 = _toString.call(state.dump);
      const inblock = block2;
      if (block2) {
        block2 = state.flowLevel < 0 || state.flowLevel > level;
      }
      const objectOrArray = type2 === "[object Object]" || type2 === "[object Array]";
      let duplicateIndex;
      let duplicate;
      if (objectOrArray) {
        duplicateIndex = state.duplicates.indexOf(object);
        duplicate = duplicateIndex !== -1;
      }
      if (state.tag !== null && state.tag !== "?" || duplicate || state.indent !== 2 && level > 0) {
        compact = false;
      }
      if (duplicate && state.usedDuplicates[duplicateIndex]) {
        state.dump = "*ref_" + duplicateIndex;
      } else {
        if (objectOrArray && duplicate && !state.usedDuplicates[duplicateIndex]) {
          state.usedDuplicates[duplicateIndex] = true;
        }
        if (type2 === "[object Object]") {
          if (block2 && Object.keys(state.dump).length !== 0) {
            writeBlockMapping(state, level, state.dump, compact);
            if (duplicate) {
              state.dump = "&ref_" + duplicateIndex + state.dump;
            }
          } else {
            writeFlowMapping(state, level, state.dump);
            if (duplicate) {
              state.dump = "&ref_" + duplicateIndex + " " + state.dump;
            }
          }
        } else if (type2 === "[object Array]") {
          if (block2 && state.dump.length !== 0) {
            if (state.noArrayIndent && !isblockseq && level > 0) {
              writeBlockSequence(state, level - 1, state.dump, compact);
            } else {
              writeBlockSequence(state, level, state.dump, compact);
            }
            if (duplicate) {
              state.dump = "&ref_" + duplicateIndex + state.dump;
            }
          } else {
            writeFlowSequence(state, level, state.dump);
            if (duplicate) {
              state.dump = "&ref_" + duplicateIndex + " " + state.dump;
            }
          }
        } else if (type2 === "[object String]") {
          if (state.tag !== "?") {
            writeScalar(state, state.dump, level, iskey, inblock);
          }
        } else if (type2 === "[object Undefined]") {
          return false;
        } else {
          if (state.skipInvalid) return false;
          throw new YAMLException2("unacceptable kind of an object to dump " + type2);
        }
        if (state.tag !== null && state.tag !== "?") {
          let tagStr = encodeURI(
            state.tag[0] === "!" ? state.tag.slice(1) : state.tag
          ).replace(/!/g, "%21");
          if (state.tag[0] === "!") {
            tagStr = "!" + tagStr;
          } else if (tagStr.slice(0, 18) === "tag:yaml.org,2002:") {
            tagStr = "!!" + tagStr.slice(18);
          } else {
            tagStr = "!<" + tagStr + ">";
          }
          state.dump = tagStr + " " + state.dump;
        }
      }
      return true;
    }
    function getDuplicateReferences(object, state) {
      const objects = [];
      const duplicatesIndexes = [];
      inspectNode(object, objects, duplicatesIndexes);
      const length = duplicatesIndexes.length;
      for (let index = 0; index < length; index += 1) {
        state.duplicates.push(objects[duplicatesIndexes[index]]);
      }
      state.usedDuplicates = new Array(length);
    }
    function inspectNode(object, objects, duplicatesIndexes) {
      if (object !== null && typeof object === "object") {
        const index = objects.indexOf(object);
        if (index !== -1) {
          if (duplicatesIndexes.indexOf(index) === -1) {
            duplicatesIndexes.push(index);
          }
        } else {
          objects.push(object);
          if (Array.isArray(object)) {
            for (let i = 0, length = object.length; i < length; i += 1) {
              inspectNode(object[i], objects, duplicatesIndexes);
            }
          } else {
            const objectKeyList = Object.keys(object);
            for (let i = 0, length = objectKeyList.length; i < length; i += 1) {
              inspectNode(object[objectKeyList[i]], objects, duplicatesIndexes);
            }
          }
        }
      }
    }
    function dump2(input, options2) {
      options2 = options2 || {};
      const state = new State(options2);
      if (!state.noRefs) getDuplicateReferences(input, state);
      let value = input;
      if (state.replacer) {
        value = state.replacer.call({ "": value }, "", value);
      }
      if (writeNode(state, 0, value, true, true)) return state.dump + "\n";
      return "";
    }
    dumper.dump = dump2;
    return dumper;
  }
  var hasRequiredJsYaml;
  function requireJsYaml() {
    if (hasRequiredJsYaml) return jsYaml;
    hasRequiredJsYaml = 1;
    const loader2 = requireLoader();
    const dumper2 = requireDumper();
    function renamed(from, to) {
      return function() {
        throw new Error("Function yaml." + from + " is removed in js-yaml 4. Use yaml." + to + " instead, which is now safe by default.");
      };
    }
    jsYaml.Type = requireType();
    jsYaml.Schema = requireSchema();
    jsYaml.FAILSAFE_SCHEMA = requireFailsafe();
    jsYaml.JSON_SCHEMA = requireJson();
    jsYaml.CORE_SCHEMA = requireCore();
    jsYaml.DEFAULT_SCHEMA = require_default();
    jsYaml.load = loader2.load;
    jsYaml.loadAll = loader2.loadAll;
    jsYaml.dump = dumper2.dump;
    jsYaml.YAMLException = requireException();
    jsYaml.types = {
      binary: requireBinary(),
      float: requireFloat(),
      map: requireMap(),
      null: require_null(),
      pairs: requirePairs(),
      set: requireSet(),
      timestamp: requireTimestamp(),
      bool: requireBool(),
      int: requireInt(),
      merge: requireMerge(),
      omap: requireOmap(),
      seq: requireSeq(),
      str: requireStr()
    };
    jsYaml.safeLoad = renamed("safeLoad", "load");
    jsYaml.safeLoadAll = renamed("safeLoadAll", "loadAll");
    jsYaml.safeDump = renamed("safeDump", "dump");
    return jsYaml;
  }
  var jsYamlExports = requireJsYaml();
  var yaml = /* @__PURE__ */ getDefaultExportFromCjs(jsYamlExports);
  var {
    Type,
    Schema,
    FAILSAFE_SCHEMA,
    JSON_SCHEMA,
    CORE_SCHEMA,
    DEFAULT_SCHEMA,
    load,
    loadAll,
    dump,
    YAMLException,
    types,
    safeLoad,
    safeLoadAll,
    safeDump
  } = yaml;

  // src/render/md.js
  marked.setOptions({ gfm: true, breaks: true, async: false });
  var PURIFY_CFG = {
    ALLOWED_TAGS: [
      "a",
      "p",
      "br",
      "hr",
      "h1",
      "h2",
      "h3",
      "h4",
      "h5",
      "h6",
      "ul",
      "ol",
      "li",
      "blockquote",
      "pre",
      "code",
      "em",
      "strong",
      "del",
      "table",
      "thead",
      "tbody",
      "tr",
      "th",
      "td",
      "span",
      "img",
      "button"
    ],
    ALLOWED_ATTR: ["href", "src", "alt", "title", "type"],
    ALLOW_DATA_ATTR: false
    // URI 白名单沿用 DOMPurify 默认（http(s)/mailto/相对——javascript:/data: 默认已封）。
  };
  function purifySupported() {
    return !!(purify_default && purify_default.isSupported);
  }
  function sanitizeHtml(html3) {
    if (!purifySupported()) return "";
    return purify_default.sanitize(String(html3 == null ? "" : html3), PURIFY_CFG);
  }
  function mdCodeBlock(text2, lang) {
    const label = String(lang || "").trim() || t("\u4EE3\u7801\u5757");
    return '<div class="md-code-block wrap"><div class="md-code-head"><span class="md-code-lang">' + esc(trunc(label, 24)) + '</span><span class="md-code-acts"><button type="button" class="md-code-btn md-code-wrap">' + esc(t("\u81EA\u52A8\u6362\u884C")) + '</button><button type="button" class="md-code-btn md-code-copy">' + esc(t("\u590D\u5236")) + '</button></span></div><pre class="md-pre"><code>' + esc(String(text2 == null ? "" : text2)) + "</code></pre></div>";
  }
  marked.use({
    renderer: {
      code(tok) {
        const t2 = tok && typeof tok === "object" ? tok : { text: tok };
        return mdCodeBlock(t2.text, t2.lang);
      }
    }
  });
  function inline2(s) {
    let out = String(s).replace(/`([^`]+)`/g, '<code class="md-code">$1</code>').replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>").replace(/(^|[^*])\*([^*\n]+)\*/g, "$1<em>$2</em>");
    out = out.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (m, text2, href) => {
      const ok = safeHref(href);
      return ok ? '<a class="md-a" href="' + esc(ok) + '">' + text2 + "</a>" : text2 + ' <span class="mono faint small">(' + href + ")</span>";
    });
    return out;
  }
  function minimalRender(src) {
    const lines = String(src == null ? "" : src).split(/\r?\n/);
    const out = [];
    const para = [];
    const flush = () => {
      if (para.length) {
        out.push('<p class="md-p">' + inline2(esc(para.join(" "))) + "</p>");
        para.length = 0;
      }
    };
    const isList = (l) => /^(\s*)([-*+]|\d+[.)])\s+/.test(l);
    let i = 0;
    while (i < lines.length) {
      const line = lines[i];
      if (!line.trim()) {
        flush();
        i += 1;
        continue;
      }
      if (/^\s*```/.test(line)) {
        flush();
        const langM = /^\s*```(\S*)/.exec(line);
        const buf = [];
        i += 1;
        while (i < lines.length && !/^\s*```/.test(lines[i])) {
          buf.push(lines[i]);
          i += 1;
        }
        i += 1;
        out.push(mdCodeBlock(buf.join("\n"), langM ? langM[1] : ""));
        continue;
      }
      const h = /^(#{1,6})\s+(.*)$/.exec(line);
      if (h) {
        flush();
        const lv = Math.min(h[1].length, 6);
        out.push("<h" + lv + ' class="md-h md-h' + lv + '">' + inline2(esc(h[2])) + "</h" + lv + ">");
        i += 1;
        continue;
      }
      if (/^\s*(-{3,}|\*{3,}|_{3,})\s*$/.test(line)) {
        flush();
        out.push('<hr class="md-hr">');
        i += 1;
        continue;
      }
      if (/^\s*>\s?/.test(line)) {
        flush();
        const q = [];
        while (i < lines.length && /^\s*>\s?/.test(lines[i])) {
          q.push(lines[i].replace(/^\s*>\s?/, ""));
          i += 1;
        }
        out.push('<blockquote class="md-quote">' + inline2(esc(q.join(" "))) + "</blockquote>");
        continue;
      }
      if (/^\s*\|.+\|\s*$/.test(line) && i + 1 < lines.length && /^\s*\|[\s:|-]+\|\s*$/.test(lines[i + 1])) {
        flush();
        const heads = line.trim().replace(/^\||\|$/g, "").split("|").map((x) => x.trim());
        i += 2;
        const body = [];
        while (i < lines.length && /^\s*\|.+\|\s*$/.test(lines[i])) {
          body.push(lines[i].trim().replace(/^\||\|$/g, "").split("|").map((x) => x.trim()));
          i += 1;
        }
        out.push('<table class="md-tbl"><thead><tr>' + heads.map((x) => "<th>" + inline2(esc(x)) + "</th>").join("") + "</tr></thead><tbody>" + body.map((row) => "<tr>" + heads.map((_, ci) => "<td>" + inline2(esc(row[ci] == null ? "" : row[ci])) + "</td>").join("") + "</tr>").join("") + "</tbody></table>");
        continue;
      }
      if (isList(line)) {
        flush();
        const items = [];
        while (i < lines.length && isList(lines[i])) {
          const m = /^(\s*)([-*+]|\d+[.)])\s+(.*)$/.exec(lines[i]);
          items.push({ depth: m[1].length >= 2 ? 1 : 0, ord: /\d/.test(m[2]), text: m[3] });
          i += 1;
        }
        let html3 = "<" + (items[0].ord ? "ol" : "ul") + ' class="md-list md-list-root">';
        const stack = [{ tag: items[0].ord ? "ol" : "ul" }];
        let openLi = false;
        for (const it of items) {
          if (openLi) {
            html3 += "</li>";
            openLi = false;
          }
          const want = it.ord ? "ol" : "ul";
          while (stack.length > it.depth + 1) html3 += "</" + stack.pop().tag + ">";
          if (stack.length === it.depth + 1 && stack[stack.length - 1].tag !== want) {
            html3 += "</" + stack.pop().tag + "><" + want + ' class="md-list">';
            stack.push({ tag: want });
          }
          while (stack.length < it.depth + 1) {
            html3 += "<" + want + ' class="md-list">';
            stack.push({ tag: want });
          }
          html3 += '<li class="md-li">' + inline2(esc(it.text));
          openLi = true;
        }
        if (openLi) html3 += "</li>";
        while (stack.length) html3 += "</" + stack.pop().tag + ">";
        out.push('<div class="md-wrap">' + html3 + "</div>");
        continue;
      }
      para.push(line.trim());
      i += 1;
    }
    flush();
    return out.join("");
  }
  function renderMarkdown(src) {
    const s = String(src == null ? "" : src);
    if (purifySupported()) {
      let raw = "";
      try {
        raw = marked.parse(s);
      } catch {
        raw = "";
      }
      const clean2 = sanitizeHtml(raw);
      return clean2 || minimalRender(s);
    }
    return minimalRender(s);
  }
  function loadYaml(src) {
    try {
      const v = yaml.load(String(src == null ? "" : src), { json: true });
      return v && typeof v === "object" ? v : null;
    } catch {
      return null;
    }
  }
  var FIELD_GLOSS = {
    name: t("\u540D\u79F0"),
    scope: t("\u4F5C\u7528\u57DF"),
    resourceVersion: t("\u7248\u672C\u53F7"),
    goalRef: t("\u76EE\u6807\u6587\u6863"),
    descRef: t("\u8BF4\u660E\u6587\u6863"),
    docs: t("\u76F8\u5173\u6587\u6863"),
    externalRefs: t("\u5916\u90E8\u4F9D\u8D56"),
    instances: t("\u670D\u52A1\u5B9E\u4F8B"),
    builtinServices: t("\u5185\u7F6E\u670D\u52A1"),
    agentRefs: t("\u5BF9\u8BDD\u5165\u53E3\uFF08Agent\uFF09"),
    phase: t("\u8FD0\u884C\u72B6\u6001"),
    observedGeneration: t("\u5DF2\u5E94\u7528\u7684\u7248\u672C"),
    labels: t("\u6807\u7B7E"),
    mounts: t("\u5DF2\u6302\u8F7D\u8FDE\u63A5"),
    connectorOps: t("\u5DF2\u6302\u8F7D\u8FDE\u63A5"),
    provider: t("\u51ED\u636E\u6E90"),
    session: t("\u5F53\u524D\u4F1A\u8BDD"),
    tags: t("\u6807\u7B7E"),
    title: t("\u6807\u9898"),
    dependsOn: t("\u4F9D\u8D56"),
    agent: t("\u8D1F\u8D23")
  };
  function cardVal(v, depth) {
    if (v == null) return '<span class="faint">\u2014</span>';
    if (Array.isArray(v)) {
      if (!v.length) return t('<span class="faint">\uFF08\u7A7A\uFF09</span>');
      return '<div class="md-chips">' + v.slice(0, 24).map((x) => '<span class="chip mute small md-chip">' + cardVal(x, depth + 1) + "</span>").join("") + (v.length > 24 ? t('<span class="faint small">\u2026\u5171 ') + v.length + t(" \u9879</span>") : "") + "</div>";
    }
    if (typeof v === "object") {
      if (depth >= 4) return t('<span class="mono faint small">\uFF08\u5D4C\u5957\u7ED3\u6784\u89C1\u300CYAML\u300D\uFF09</span>');
      const keys = Object.keys(v);
      if (!keys.length) return t('<span class="faint">\uFF08\u7A7A\uFF09</span>');
      return '<div class="md-subcards">' + keys.map(
        (k) => '<div class="md-subcard"><div class="md-subcard-k">' + esc(k) + '</div><div class="md-subcard-v">' + cardVal(v[k], depth + 1) + "</div></div>"
      ).join("") + "</div>";
    }
    const s = String(v);
    if (/^(agent|system|session):/.test(s)) return '<span class="lk md-lk">' + esc(s) + "</span>";
    if (/^https?:\/\//i.test(s)) {
      const href = safeHref(s);
      return href ? '<a class="md-a" href="' + esc(href) + '">' + esc(trunc(s, 80)) + "</a>" : esc(trunc(s, 80));
    }
    return esc(trunc(s, 300));
  }
  function cardHtml(key, v, depth) {
    const g = FIELD_GLOSS[String(key)];
    return '<div class="md-card"><div class="md-card-k">' + esc(String(key)) + (g ? '<span class="md-gloss">' + esc(g) + "</span>" : "") + '</div><div class="md-card-v">' + cardVal(v, depth) + "</div></div>";
  }
  function yamlToCards(src) {
    const tree = loadYaml(src);
    if (!tree || !Object.keys(tree).length) return "";
    return '<div class="md-cards">' + Object.keys(tree).map((k) => cardHtml(k, tree[k], 0)).join("") + "</div>";
  }
  function rawDisclosure(title, content) {
    return '<details class="md-raw"><summary>' + esc(title || t("\u539F\u6587")) + '</summary><pre class="mono md-raw-pre">' + esc(content == null ? "" : content) + "</pre></details>";
  }
  function yamlTextView(content, title) {
    const s = String(content == null ? "" : content);
    const rows = Math.min(Math.max(s.split("\n").length + 1, 4), 48);
    return (title ? '<div class="small faint" style="margin-bottom:4px">' + esc(title) + "</div>" : "") + '<textarea class="inp mono" readonly rows="' + rows + '" spellcheck="false" style="width:100%;font-size:12px;line-height:1.5;resize:vertical;white-space:pre">' + esc(s) + "</textarea>";
  }
  function dumpYaml(value) {
    if (value == null) return "";
    try {
      const s = yaml.dump(value, { lineWidth: 240, noRefs: true });
      return typeof s === "string" ? s : "";
    } catch {
      return "";
    }
  }
  function splitFrontmatter(src) {
    const s = String(src == null ? "" : src);
    if (!s.startsWith("---")) return { fm: "", body: s };
    const end = s.indexOf("\n---", 3);
    if (end < 0) return { fm: "", body: s };
    return { fm: s.slice(4, end), body: s.slice(end + 4).replace(/^\n/, "") };
  }

  // src/screens/workflow-config.js
  function defaultLoopSpec(name = "core-system-developer-loop", existing = {}) {
    const s = existing && typeof existing === "object" ? existing : {};
    const logSrc = s.logSource || {};
    const collector = logSrc.collector || {};
    const timer = logSrc.timer || s.triggers && s.triggers.timer || {};
    const approval = s.approval || {};
    const sysBind = s.systemBindings || {};
    const devSys = sysBind.developerSystem || {};
    const depSys = sysBind.deploySystem || {};
    const devConv = devSys.conventions || {};
    const ver = s.verification || {};
    const reg = s.regression || {};
    const llm = s.llm || {};
    const defaultConditions = [
      { id: "lint-rules", kind: "code", title: "\u4EE3\u7801 Lint \u4E0E\u89C4\u5219\u5BA1\u6838 (R-PUB-1 / R-DOC-6 / R-ENG-7 \u2264800\u884C)", command: "connectops-code-audit/audit", required: true },
      { id: "unit-tests", kind: "test", title: "\u540E\u7AEF\u4E0E\u524D\u7AEF\u5355\u5143/\u96C6\u6210\u6D4B\u8BD5\u5168\u7EFF (go test ./... & npm test)", command: "go test ./... && npm test", required: true },
      { id: "live-probe", kind: "agent", title: "\u6D3B\u4F53\u63A2\u9488\u4E0E\u9875\u9762\u65E0\u5047\u6570\u636E/\u65E0\u62A5\u9519\u9A8C\u8BC1", command: "agent-verify", required: true }
    ];
    const defaultPostActions = [
      { id: "bugdoc-lifecycle", kind: "doc-lifecycle", targetDocType: "BugDoc", action: reg.bugDocPolicy || "invalidate", enabled: true, title: "\u4E34\u65F6 BugDoc \u6536\u5C3E\u5904\u7406\uFF08\u6807\u8BB0\u5931\u6548 invalidate / \u5220\u9664 delete / \u4FDD\u7559 keep\uFF0C\u53EF\u589E\u5220\u6B64\u9879\uFF09" },
      { id: "status-insync", kind: "status-writeback", targetPhase: "InSync", enabled: true, title: "\u56DE\u5199 ServiceInstance.status.phase = InSync\uFF08\u6E05\u96F6\u6F02\u79FB\uFF09" }
    ];
    const defaultPhases = [
      { name: "log-event", phase: "Idle", title: "1. \u76D1\u542C\u65E5\u5FD7\u4E0E Event", unit: "connectops", connector: collector.connector || "connectops-log-collector", op: collector.op || "collect-logs", provider: llm.provider || "openai-default", modelTier: "low", entryEvents: ["BugReport", "MetricThreshold", "ExternalChanged", "ReconcileTick", "LoopRequested"] },
      { name: "agent-judge", phase: "Analyzing", title: "2. Agent \u5224\u65AD\uFF08\u8BC4\u4F30\u4FEE\u590D\u4EF7\u503C/\u4EA7\u51FA\u65B9\u6848\u6216 BugDoc\uFF09", unit: "agent", op: "judge-issue", provider: llm.provider || "openai-default", modelTier: "high", refRules: ["rule/rule_eng", "rule/rule_doc"] },
      { name: "approval-gate", phase: "AwaitingApproval", title: "3. \u5BA1\u6279 / \u81EA\u52A8\u5224\u65AD\u5BA1\u6279\uFF08\u63A7\u5236\u8282\u70B9\uFF09", unit: "gate", op: "approval-check" },
      { name: "doc-apply", phase: "Planning", title: "4. Doc Apply\uFF08\u5BF9\u8BDD\u6539\u6587\u6863/BugDoc \u7B2C\u4E8C\u5207\u5165\u70B9\uFF09", unit: "agent", op: "apply-doc", provider: llm.provider || "openai-default", modelTier: "standard", refDocs: ["goal"], refRules: ["rule/rule_doc", "rule/rule_eng"], refTags: ["audit", "fix"], tagRules: ["rule/tag_default"], entryEvents: ["DocApplied"] },
      { name: "developer-system", phase: "Executing", title: "5. developer-system \u7814\u53D1\uFF08service \u5185\u90E8 agent \u6267\u884C\uFF09", unit: "service", connector: devSys.connector || "connectops-developer", op: devSys.op || "develop", service: devSys.service || "developer-system", contextPassing: { previousRound: true, fields: ["judge", "plan", "diff", "approval"] }, permissionRole: "role/developer-system-ops" },
      { name: "verify", phase: "Verifying", title: "6. \u9A8C\u6536\uFF08developer-system Agent \u9A8C\u6536 session\uFF09", unit: "service", service: devSys.service || "developer-system", op: "acceptance", contextPassing: { previousRound: true, fields: ["diff", "verifyConditions"] } },
      { name: "deploy-system", phase: "Deploying", title: "7. deploy-system \u53D1\u7248\uFF08service \u5185\u90E8 agent \u6267\u884C\uFF09", unit: "service", connector: depSys.connector || "connectops-deploy", op: depSys.op || "release", service: depSys.service || "deploy-system", contextPassing: { previousRound: true, fields: ["verifyReport", "release"] }, permissionRole: "role/deploy-system-ops" },
      { name: "report-archive", phase: "Reporting", title: "8. \u603B\u7ED3\u5F52\u6863\uFF08\u89E3\u91CA\u6267\u884C postActions + \u56DE\u5199 InSync\uFF09", unit: "agent", op: "archive-report", provider: llm.provider || "openai-default", modelTier: "low", tools: ["findings-read", "doc-draft", "tag-write"] }
    ];
    const defaultTransitions = [
      { from: "Idle", on: "advance", to: "Analyzing" },
      { from: "Idle", on: "entry:DocApplied", to: "Planning" },
      { from: "Analyzing", on: "advance", to: "AwaitingApproval" },
      { from: "Analyzing", on: "worth_fixing", to: "AwaitingApproval" },
      { from: "Analyzing", on: "ignore", to: "Reporting" },
      { from: "AwaitingApproval", on: "advance", to: "Planning" },
      { from: "AwaitingApproval", on: "ApprovalGranted", to: "Planning" },
      { from: "AwaitingApproval", on: "ApprovalRejected", to: "Failed" },
      { from: "Planning", on: "advance", to: "Executing" },
      { from: "Executing", on: "advance", to: "Verifying" },
      { from: "Verifying", on: "advance", to: "Deploying" },
      { from: "Verifying", on: "retry-develop", to: "Executing" },
      { from: "Deploying", on: "advance", to: "Reporting" }
    ];
    return {
      name,
      enabled: s.enabled != null ? Boolean(s.enabled) : false,
      immutableNodes: Boolean(s.immutableNodes),
      // 2026-11-27 用户裁决：节点可加删（内置主干也不恒锁定）——前端节点表加节点/删节点/改参数，draft→apply 落账
      undeletable: name === "core-system-developer-loop" || name === "default-task" ? true : Boolean(s.undeletable),
      servicesInvolved: Array.isArray(s.services && s.services.involved) ? s.services.involved.join(", ") : "self-recursive-core, developer-system, deploy-system",
      servicesMode: s.services && s.services.mode || "act",
      provider: llm.provider || "openai-default",
      defaultModel: llm.model || "reasoning-tier",
      globalRuleRefs: Array.isArray(s.globalRuleRefs) ? s.globalRuleRefs.join(", ") : "rule/rule_eng, rule/rule_doc, rule/rule_ux, rule/rule_ctx, rule/rule_pub",
      logEvents: Array.isArray(logSrc.events) ? logSrc.events.join(", ") : "BugReport, MetricThreshold, ExternalChanged, ReconcileTick, LoopRequested",
      logMinLevel: logSrc.minLevel || "warn",
      logServices: Array.isArray(logSrc.services) ? logSrc.services.join(", ") : "self-recursive-core",
      logKeyword: logSrc.keywordMatch || "ERROR|PANIC|FAIL|drift|unhealthy",
      timerEnabled: timer.enabled != null ? Boolean(timer.enabled) : true,
      timerInterval: timer.interval || timer.resync || "30m",
      timerCron: timer.cron || "*/30 * * * *",
      collectorConnector: collector.connector || "connectops-log-collector",
      collectorOp: collector.op || "collect-logs",
      collectorQuery: collector.query || "level>=warn",
      docApplyEntryEnabled: true,
      docApplyDocTypes: "Goal, Desc, Rule, WorkflowDesign, ImplDoc, VerifyDoc, BugDoc",
      approvalMode: approval.mode || "agent",
      approvalCriteria: approval.autoCriteria || "\u5355\u6D4B\u5168\u7EFF\u3001\u7B26\u5408\u5168\u5C40 rule_eng/rule_doc \u89C4\u8303\u4E14\u4E0D\u6D89\u53CA\u7834\u574F\u6027\u9AD8\u5371\u5220\u5E93\u64CD\u4F5C\u65F6\u81EA\u52A8\u6279\u51C6\u653E\u884C",
      approvalMaxRisk: approval.maxRiskLevel || "low",
      devService: devSys.service || "developer-system",
      devResolution: devSys.resolution || "default",
      devConnector: devSys.connector || "connectops-developer",
      devOp: devSys.op || "develop",
      devBranchPrefix: devConv.branchPrefix || "agent/",
      devWorktreeDir: devConv.worktreeDir || "worktrees/",
      devMaxLines: devConv.maxFileLines || 800,
      deployService: depSys.service || "deploy-system",
      deployResolution: depSys.resolution || "default",
      deployConnector: depSys.connector || "connectops-deploy",
      deployOp: depSys.op || "release",
      verifyMode: ver.mode || "both",
      verifyConnector: ver.connector || "connectops-code-audit",
      verifyOp: ver.op || "audit",
      verifyOnFail: ver.onFail || "retry-develop",
      conditions: Array.isArray(ver.conditions) && ver.conditions.length ? ver.conditions : defaultConditions,
      bugDocPolicy: reg.bugDocPolicy || (reg.cleanTempBugDoc === true ? "delete" : "invalidate"),
      statusWriteback: reg.statusWriteback != null ? Boolean(reg.statusWriteback) : true,
      postActions: Array.isArray(reg.postActions) && reg.postActions.length ? reg.postActions : defaultPostActions,
      phases: Array.isArray(s.phases) && s.phases.length && typeof s.phases[0] === "object" ? s.phases.map((n) => n && n.unit === "session" ? { ...n, unit: "agent" } : n) : defaultPhases,
      transitions: Array.isArray(s.transitions) && s.transitions.length ? s.transitions.map((t2) => ({ from: t2 && t2.from, on: t2 && t2.on, to: t2 && t2.to })).filter((t2) => t2.from && t2.to) : defaultTransitions
    };
  }
  function buildWorkflowYaml(cfg, scope = "global") {
    const splitList = (v) => String(v || "").split(",").map((x) => x.trim()).filter(Boolean);
    const inv = splitList(cfg.servicesInvolved);
    const rules = splitList(cfg.globalRuleRefs);
    const logEvs = splitList(cfg.logEvents);
    const logSvcs = splitList(cfg.logServices);
    const docTypes = splitList(cfg.docApplyDocTypes);
    const q = (s) => JSON.stringify(String(s == null ? "" : s));
    const condLines = (cfg.conditions || []).map(
      (c, idx) => `      - {id: ${q(c.id || "cond-" + (idx + 1))}, kind: ${q(c.kind || "code")}, title: ${q(c.title || "")}, command: ${q(c.command || "")}, required: ${Boolean(c.required)}}`
    ).join("\n");
    const actionLines = (cfg.postActions || []).map(
      (a, idx) => `      - {id: ${q(a.id || "act-" + (idx + 1))}, kind: ${q(a.kind || "doc-lifecycle")}, targetDocType: ${q(a.targetDocType || "BugDoc")}, action: ${q(a.action || cfg.bugDocPolicy || "invalidate")}, enabled: ${Boolean(a.enabled)}, title: ${q(a.title || "")}}`
    ).join("\n");
    const nodeLine = (n) => {
      const p = [`name: ${n.name || "node"}`, `phase: ${n.phase || "Idle"}`, `title: ${q(n.title || n.name || "")}`, `unit: ${n.unit || "connectops"}`];
      if (n.connector) p.push(`connector: ${n.connector}`);
      if (n.op) p.push(`op: ${n.op}`);
      if (n.service) p.push(`service: ${n.service}`);
      if (n.provider) p.push(`provider: ${n.provider}`);
      if (n.model) p.push(`model: ${q(n.model)}`);
      if (n.modelTier) p.push(`modelTier: ${n.modelTier}`);
      if (n.permissionRole) p.push(`permissionRole: ${n.permissionRole}`);
      ["refDocs", "refRules", "refTags", "tagRules", "tools", "entryEvents"].forEach((k) => {
        if (Array.isArray(n[k]) && n[k].length) p.push(`${k}: [${n[k].map(q).join(", ")}]`);
      });
      if (n.contextPassing && (n.contextPassing.previousRound || Array.isArray(n.contextPassing.fields) && n.contextPassing.fields.length)) {
        p.push(`contextPassing: {previousRound: ${n.contextPassing.previousRound !== false}, fields: [${(n.contextPassing.fields || []).map(q).join(", ")}]}`);
      }
      return `    - {${p.join(", ")}}`;
    };
    const phaseLines = (cfg.phases || []).map(nodeLine).join("\n");
    const nodeMappingLines = (cfg.phases || []).map(
      (n) => `      ${n.name}: {tier: ${n.modelTier || "standard"}${n.model ? ", model: " + q(n.model) : ""}}`
    ).join("\n");
    const transLines = (cfg.transitions || []).map(
      (t2) => `    - {from: ${t2.from}, on: ${q(String(t2.on || "advance"))}, to: ${t2.to}}`
    ).join("\n");
    return `apiVersion: core/v1
kind: Workflow
metadata:
  name: ${cfg.name}
  scope: ${scope}
  labels:
    core/origin: builtin
spec:
  enabled: ${Boolean(cfg.enabled)}
  immutableNodes: ${Boolean(cfg.immutableNodes)}
  undeletable: ${Boolean(cfg.undeletable)}
  connector: connector-ops/${cfg.name}
  services:
    involved: [${inv.map(q).join(", ")}]
    mode: ${cfg.servicesMode || "act"}
  llm:
    provider: ${cfg.provider || "openai-default"}
    model: ${q(cfg.defaultModel || "reasoning-tier")}
    tier: high
    nodeMapping:
${nodeMappingLines}
  globalRuleRefs: [${rules.map(q).join(", ")}]
  logSource:
    events: [${logEvs.map(q).join(", ")}]
    minLevel: ${cfg.logMinLevel || "warn"}
    services: [${logSvcs.map(q).join(", ")}]
    keywordMatch: ${q(cfg.logKeyword || "")}
    timer:
      enabled: ${Boolean(cfg.timerEnabled)}
      interval: ${q(cfg.timerInterval || "30m")}
      cron: ${q(cfg.timerCron || "*/30 * * * *")}
    collector:
      connector: ${cfg.collectorConnector || "connectops-log-collector"}
      op: ${cfg.collectorOp || "collect-logs"}
      query: ${q(cfg.collectorQuery || "level>=warn")}
      autoDevelopOnMissing: true
  triggers:
    events:
      types: [${logEvs.concat(cfg.docApplyEntryEnabled ? ["DocApplied"] : []).map(q).join(", ")}]
      minLevel: ${cfg.logMinLevel || "warn"}
    timer:
      resync: ${q(cfg.timerInterval || "30m")}
      enabled: ${Boolean(cfg.timerEnabled)}
      interval: ${q(cfg.timerInterval || "30m")}
      cron: ${q(cfg.timerCron || "*/30 * * * *")}
  entryPoints:
    - {node: log-event, phase: Idle, enabled: true, events: [${logEvs.map(q).join(", ")}], description: "\u5165\u53E3 1\uFF1A\u76D1\u542C\u65E5\u5FD7/\u8FD0\u884C\u671F Bug/\u5B9A\u65F6\u5DE1\u68C0\u4E8B\u4EF6\u4ECE\u5934\u6D41\u8F6C"}
    - {node: doc-apply, phase: Planning, enabled: ${Boolean(cfg.docApplyEntryEnabled)}, events: ["DocApplied"], docTypes: [${docTypes.map(q).join(", ")}], description: "\u5165\u53E3 2\uFF1A\u7528\u6237\u5BF9\u8BDD\u4FEE\u6539\u6587\u6863\u6216\u4EA7\u51FA\u4E34\u65F6 BugDoc \u89E6\u53D1 DocApplied \u76F4\u63A5\u4ECE\u8282\u70B9 4 \u5207\u5165"}
  ingest:
    dedup: {window: 5m}
    group: {by: service, max: 20}
    filterLevels: [warn, error, fatal]
  approval:
    mode: ${cfg.approvalMode || "agent"}
    autoCriteria: ${q(cfg.approvalCriteria || "")}
    maxRiskLevel: ${cfg.approvalMaxRisk || "low"}
  systemBindings:
    developerSystem:
      service: ${cfg.devService || "developer-system"}
      resolution: ${cfg.devResolution || "default"}
      connector: ${cfg.devConnector || "connectops-developer"}
      op: ${cfg.devOp || "develop"}
      conventions:
        branchPrefix: ${q(cfg.devBranchPrefix || "agent/")}
        worktreeDir: ${q(cfg.devWorktreeDir || "worktrees/")}
        tmpDir: "tmp/"
        maxFileLines: ${Number(cfg.devMaxLines) || 800}
    deploySystem:
      service: ${cfg.deployService || "deploy-system"}
      resolution: ${cfg.deployResolution || "default"}
      connector: ${cfg.deployConnector || "connectops-deploy"}
      op: ${cfg.deployOp || "release"}
  verification:
    mode: ${cfg.verifyMode || "both"}
    connector: ${cfg.verifyConnector || "connectops-code-audit"}
    op: ${cfg.verifyOp || "audit"}
    onFail: ${cfg.verifyOnFail || "retry-develop"}
    conditions:
${condLines || "      []"}
  phases:
${phaseLines}
  transitions:
${transLines}
  onEscalate: {action: pause, notify: bind-role}
  onFailed:
    backoff: 30s*2^n
    maxAttempts: 3
    rollbackPolicy: by-step
    escalate: {to: root, notify: bind-role}
  brakes:
    cooldown: 2m
    maxMutationsPerTick: 10
    avoidRepeatFix: true
  regression:
    statusWriteback: ${Boolean(cfg.statusWriteback)}
    bugDocPolicy: ${cfg.bugDocPolicy || "invalidate"}
    cleanTempBugDoc: ${cfg.bugDocPolicy === "delete"}
    targetPhaseOnConverge: InSync
    postActions:
${actionLines || "      []"}
    reports: [DriftReport, SessionLog]
`;
  }
  var NODE_UNITS = ["connectops", "agent", "service", "gate"];
  var UNIT_CHIP = { connectops: "mute", agent: "acc", service: "warn", gate: "llm", side: "esc" };
  function nodeRowHtml(n, i, defProvider) {
    return `<tr data-node-edit="${esc(n.name || "")}" style="cursor:pointer" title="${esc(t("\u70B9\u51FB\u5F39\u7A97\u7F16\u8F91\u8282\u70B9\u7EC6\u8282\u914D\u7F6E"))}"><td class="mono small">${i + 1}. ${esc(n.name || "")}</td><td>${chip(n.phase || "\u2014", "acc")}</td><td><b>${esc(n.title || "")}</b></td><td class="mono small">${chip(n.unit === "gate" ? t("\u63A7\u5236\u8282\u70B9") : n.unit || "connectops", UNIT_CHIP[n.unit] || "mute")}</td><td class="mono small">${n.connector ? `<span class="lk" data-open-conn="${esc(n.connector)}">${esc(n.connector)}</span>` : "\u2014"}${n.op ? ` \xB7 ${esc(n.op)}` : ""}</td><td><span class="chip acc tiny click" data-open-provider="${esc(n.provider || defProvider)}" title="${esc(t("\u70B9\u51FB\u914D\u7F6E OpenAIProvider \u8D26\u53F7/baseUrl/\u6A21\u578B\u6620\u5C04"))}">Provider: ${esc(n.provider || defProvider)}</span> ${chip("tier:" + (n.modelTier || "standard"), n.modelTier === "high" ? "warn" : "mute")}</td><td class="small mono">${Array.isArray(n.refRules) && n.refRules.length ? esc(n.refRules.join(", ")) : "\u2014"}${Array.isArray(n.refTags) && n.refTags.length ? ' <span class="faint">tag:</span>' + esc(n.refTags.join(",")) : ""}</td></tr>`;
  }
  function nodeFieldsHtml(n, defProvider) {
    const u = n.unit || "connectops";
    const grid = (inner) => `<div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:6px;margin-top:6px">${inner}</div>`;
    const inp = (lab, key, val) => `<div><div class="small faint">${esc(lab)}</div><input class="inp mono" data-nm="${key}" value="${esc(val == null ? "" : String(val))}" style="width:100%"></div>`;
    const base = grid(
      inp("\u8282\u70B9\u540D (name)", "name", n.name || "") + inp("\u76F8\u4F4D (phase)", "phase", n.phase || "Idle") + `<div><div class="small faint">${esc(t("\u804C\u8D23\u8BF4\u660E (title)"))}</div><input class="inp" data-nm="title" value="${esc(n.title || "")}" style="width:100%"></div><div><div class="small faint">${esc(t("\u6267\u884C\u5355\u5143 (unit)"))}</div><select class="inp" data-nm="unit" style="width:100%">${NODE_UNITS.map((x) => `<option value="${x}" ${u === x ? "selected" : ""}>${x}${x === "gate" ? t("\uFF08\u63A7\u5236\u8282\u70B9\uFF09") : ""}</option>`).join("")}</select></div>`
    );
    const rows = [base];
    if (u === "agent") {
      rows.push(grid(
        inp("Provider", "provider", n.provider || defProvider || "openai-default") + inp("\u6A21\u578B\u6863\u4F4D (modelTier low/standard/high)", "modelTier", n.modelTier || "standard") + inp("ref \u6587\u6863 (refDocs \u9017\u53F7\u5206\u9694)", "refDocs", (n.refDocs || []).join(", ")) + inp("ref \u89C4\u5219 (refRules)", "refRules", (n.refRules || []).join(", ")) + inp("\u9ED8\u8BA4 ref tag (refTags)", "refTags", (n.refTags || []).join(", ")) + inp("tag \u89C4\u5219 (tagRules)", "tagRules", (n.tagRules || []).join(", ")) + inp("\u968F\u8EAB\u5DE5\u5177 (tools)", "tools", (n.tools || []).join(", ")) + inp("\u5165\u53E3\u4E8B\u4EF6 (entryEvents)", "entryEvents", (n.entryEvents || []).join(", "))
      ));
    } else if (u === "service") {
      rows.push(grid(
        inp("\u76EE\u6807\u7CFB\u7EDF (service)", "service", n.service || "") + inp("\u7B97\u5B50 (op)", "op", n.op || "") + inp("ConnectorOps", "connector", n.connector || "") + inp("\u6743\u9650 Role (permissionRole\u2014\u2014P7 \u968F\u6587\u6863\u6388\u4E88)", "permissionRole", n.permissionRole || "") + inp("\u4E0A\u4E00\u8F6E\u4E0A\u4E0B\u6587\u5B57\u6BB5 (contextPassing.fields)", "ctxFields", (n.contextPassing && n.contextPassing.fields || []).join(", ")) + `<div style="display:flex;align-items:end;padding-bottom:4px"><label class="small"><input type="checkbox" data-nm="ctxPrev" ${!n.contextPassing || n.contextPassing.previousRound !== false ? "checked" : ""}> ${esc(t("\u4F20\u4E0A\u4E00\u8F6E\u8BE6\u7EC6\u4E0A\u4E0B\u6587"))}</label></div>`
      ));
    } else if (u === "connectops") {
      rows.push(grid(
        inp("ConnectorOps", "connector", n.connector || "") + inp("\u7B97\u5B50 (op)", "op", n.op || "") + inp("Provider", "provider", n.provider || defProvider || "openai-default") + inp("\u6A21\u578B\u6863\u4F4D (modelTier)", "modelTier", n.modelTier || "low") + inp("\u5165\u53E3\u4E8B\u4EF6 (entryEvents)", "entryEvents", (n.entryEvents || []).join(", "))
      ));
    } else {
      rows.push(grid(inp("\u95F8\u7B97\u5B50 (op)", "op", n.op || "approval-check")));
    }
    return rows.join("");
  }
  function readNodeFields(root, base) {
    const get = (k) => root.querySelector(`[data-nm="${k}"]`);
    const lst = (k) => String(get(k) ? get(k).value : "").split(/[，,]+/).map((x) => x.trim()).filter(Boolean);
    const u = get("unit") ? get("unit").value : base.unit || "connectops";
    const n = {
      ...base,
      name: get("name") && get("name").value.trim() || base.name || "node",
      phase: get("phase") && get("phase").value.trim() || "Idle",
      title: get("title") ? get("title").value : base.title || "",
      unit: u
    };
    if (u === "agent") {
      n.provider = get("provider") ? get("provider").value.trim() : "";
      n.modelTier = get("modelTier") ? get("modelTier").value.trim() : "standard";
      ["refDocs", "refRules", "refTags", "tagRules", "tools", "entryEvents"].forEach((k) => {
        n[k] = lst(k);
      });
      delete n.service;
      delete n.contextPassing;
      delete n.permissionRole;
    } else if (u === "service") {
      n.service = get("service") ? get("service").value.trim() : "";
      n.op = get("op") ? get("op").value.trim() : "";
      n.connector = get("connector") ? get("connector").value.trim() : "";
      n.permissionRole = get("permissionRole") ? get("permissionRole").value.trim() : "";
      n.contextPassing = { previousRound: Boolean(get("ctxPrev") && get("ctxPrev").checked), fields: lst("ctxFields") };
      ["refDocs", "refRules", "refTags", "tagRules", "tools"].forEach((k) => delete n[k]);
    } else if (u === "connectops") {
      n.connector = get("connector") ? get("connector").value.trim() : "";
      n.op = get("op") ? get("op").value.trim() : "";
      n.provider = get("provider") ? get("provider").value.trim() : "";
      n.modelTier = get("modelTier") ? get("modelTier").value.trim() : "low";
      n.entryEvents = lst("entryEvents");
      ["refDocs", "refRules", "refTags", "tagRules", "tools", "service", "contextPassing", "permissionRole"].forEach((k) => delete n[k]);
    } else {
      n.op = get("op") ? get("op").value.trim() : "approval-check";
      ["refDocs", "refRules", "refTags", "tagRules", "tools", "service", "contextPassing", "permissionRole", "connector", "provider", "modelTier", "entryEvents"].forEach((k) => delete n[k]);
    }
    return n;
  }
  function wfGraphHtml(phases, transitions) {
    const nodes = [];
    const seenKeys = /* @__PURE__ */ new Set();
    const phaseFirst = /* @__PURE__ */ new Map();
    (phases || []).forEach((n) => {
      let name, phase;
      if (typeof n === "string") {
        name = n;
        phase = n;
      } else if (!n || !n.phase && !n.name) return;
      else {
        name = String(n.name || "");
        phase = String(n.phase || n.name || "");
      }
      if (!phase) return;
      let key = name && !seenKeys.has(name) ? name : phase + "#" + nodes.length;
      while (seenKeys.has(key)) key = key + "#";
      seenKeys.add(key);
      nodes.push({ key, name: name || phase, phase, title: typeof n === "object" ? n.title || "" : "", unit: typeof n === "object" && n.unit ? n.unit : "connectops", entry: typeof n === "object" && Array.isArray(n.entryEvents) && n.entryEvents.length > 0 });
      if (!phaseFirst.has(phase)) phaseFirst.set(phase, key);
    });
    const edges = [];
    (transitions || []).forEach((t2) => {
      if (!t2 || !t2.from || !t2.to || t2.from === t2.to) return;
      [t2.from, t2.to].forEach((k) => {
        if (!phaseFirst.has(k)) {
          const key = k + "#side";
          if (!seenKeys.has(key)) {
            nodes.push({ key, name: k, phase: k, title: t("\u4FA7\u8F68\u6001\uFF08\u975E\u6B63\u5411\u4E3B\u7EBF\uFF09"), unit: "side", entry: false });
            seenKeys.add(key);
            phaseFirst.set(k, key);
          }
        }
      });
      edges.push({ from: t2.from, on: String(t2.on || "advance"), to: t2.to });
    });
    const branchPairs = new Set(edges.filter((e) => e.on !== "advance").map((e) => e.from + "\u2192" + e.to));
    const drawn = [];
    const seenPair = /* @__PURE__ */ new Set();
    edges.forEach((e) => {
      const pair = e.from + "\u2192" + e.to;
      if (e.on === "advance" && branchPairs.has(pair)) return;
      if (seenPair.has(pair + "|" + e.on)) return;
      seenPair.add(pair + "|" + e.on);
      drawn.push(e);
    });
    if (!nodes.length) return "";
    const phaseLevel = /* @__PURE__ */ new Map();
    phaseFirst.forEach((k, ph) => phaseLevel.set(ph, 0));
    for (let pass = 0; pass < nodes.length; pass++) {
      let changed = false;
      drawn.forEach((e) => {
        const lf = phaseLevel.get(e.from) || 0;
        const lt = phaseLevel.get(e.to) || 0;
        if (lt < lf + 1 && lf + 1 < nodes.length) {
          phaseLevel.set(e.to, lf + 1);
          changed = true;
        }
      });
      if (!changed) break;
    }
    let entryNo = 0;
    nodes.forEach((n) => {
      if (n.entry) n.entryLabel = t("\u5165\u53E3") + ++entryNo;
    });
    const cols = /* @__PURE__ */ new Map();
    const nodeLevel = /* @__PURE__ */ new Map();
    nodes.forEach((n) => {
      const lv = phaseLevel.get(n.phase) || 0;
      nodeLevel.set(n.key, lv);
      if (!cols.has(lv)) cols.set(lv, []);
      cols.get(lv).push(n);
    });
    const NW = 190, NH = 56, GX = 108, GY = 28;
    let maxRows = 1;
    cols.forEach((rows) => {
      maxRows = Math.max(maxRows, rows.length);
    });
    const colKeys = [...cols.keys()].sort((a, b) => a - b);
    const W = colKeys.length * (NW + GX) - GX + 8;
    const H = maxRows * (NH + GY) - GY + 10;
    const pos = /* @__PURE__ */ new Map();
    colKeys.forEach((lv, ci) => {
      const rows = cols.get(lv);
      const yOff = Math.max(5, (H - (rows.length * (NH + GY) - GY)) / 2);
      rows.forEach((n, ri) => pos.set(n.key, { x: ci * (NW + GX) + 4, y: yOff + ri * (NH + GY) }));
    });
    const nodeDiv = (n) => {
      const p = pos.get(n.key);
      const unitCls = UNIT_CHIP[n.unit] || "mute";
      const unitLabel = n.unit === "gate" ? t("\u63A7\u5236\u8282\u70B9") : n.unit === "side" ? t("\u4FA7\u8F68") : n.unit;
      return '<div class="wf-graph-node' + (n.unit === "side" ? " wf-graph-side" : "") + '" style="left:' + p.x + "px;top:" + p.y + "px;width:" + NW + "px;height:" + NH + 'px" data-graph-node="' + esc(n.name) + '" title="' + esc((n.title || n.name) + " \xB7 " + n.key) + '"><div class="wf-graph-unit"><span class="chip ' + unitCls + ' tiny">' + esc(unitLabel) + "</span>" + (n.entryLabel ? ' <span class="chip acc tiny">' + esc(n.entryLabel) + "</span>" : "") + '</div><div class="wf-graph-name mono">' + esc(n.name) + ' <span class="faint small">' + esc(n.key) + "</span></div></div>";
    };
    const svg2 = drawn.map((e) => {
      const a = pos.get(phaseFirst.get(e.from)), b = pos.get(phaseFirst.get(e.to));
      if (!a || !b) return "";
      const x1 = a.x + NW, y1 = a.y + NH / 2, x2 = b.x, y2 = b.y + NH / 2;
      const mx = (x1 + x2) / 2;
      const d = "M " + x1 + " " + y1 + " C " + mx + " " + y1 + ", " + mx + " " + y2 + ", " + x2 + " " + y2;
      const branch = !(e.on === "advance");
      const label = branch ? '<div class="wf-graph-elabel" style="left:' + mx + "px;top:" + Math.round((y1 + y2) / 2 - 9) + 'px">' + esc(e.on) + "</div>" : "";
      return '<path d="' + d + '" class="wf-graph-edge' + (branch ? " wf-graph-branch" : "") + '"></path>' + label;
    }).join("");
    return '<div class="wf-graph" style="width:' + W + "px;height:" + H + 'px"><svg width="' + W + '" height="' + H + '" style="position:absolute;left:0;top:0;pointer-events:none"><defs><marker id="wfg-arrow" markerWidth="8" markerHeight="8" refX="7" refY="3" orient="auto"><path d="M0,0 L7,3 L0,6 Z" fill="var(--border-2)"></path></marker></defs>' + svg2 + "</svg>" + nodes.map(nodeDiv).join("") + "</div>";
  }
  function workflowStudioHtml(wf, usedName) {
    const cfg = defaultLoopSpec(usedName, wf && wf.spec);
    const scope = wf && wf.primaryScope || "global";
    const yamlText = wf && wf.rawYaml ? String(wf.rawYaml) : buildWorkflowYaml(cfg, scope);
    const condRowsHtml = (cfg.conditions || []).map(
      (c, idx) => `<div class="mini-row" data-wf-cond-row="${idx}" style="display:flex;gap:8px;align-items:center;flex-wrap:wrap"><input type="checkbox" data-cond-req="${idx}" ${c.required !== false ? "checked" : ""} title="\u5FC5\u8FC7\u786C\u95E8" /><select class="inp" data-cond-kind="${idx}" style="width:110px">` + ["code", "test", "agent", "probe", "custom"].map((k) => `<option value="${k}" ${c.kind === k ? "selected" : ""}>${k}</option>`).join("") + `</select><input class="inp" data-cond-title="${idx}" value="${esc(c.title || "")}" placeholder="\u9A8C\u6536\u6761\u4EF6\u63CF\u8FF0" style="flex:2;min-width:200px" /><input class="inp mono" data-cond-cmd="${idx}" value="${esc(c.command || "")}" placeholder="\u6267\u884C\u547D\u4EE4\u6216 ConnectorOps/op" style="flex:1;min-width:160px" /><button type="button" class="btn tiny mute" data-cond-del="${idx}">\u5220\u9664</button></div>`
    ).join("");
    const postActRowsHtml = (cfg.postActions || []).map(
      (a, idx) => `<div class="mini-row" data-wf-act-row="${idx}" style="display:flex;gap:8px;align-items:center;flex-wrap:wrap"><input type="checkbox" data-act-en="${idx}" ${a.enabled !== false ? "checked" : ""} title="\u542F\u7528\u6B64\u6536\u5C3E\u52A8\u4F5C" /><select class="inp" data-act-kind="${idx}" style="width:140px">` + ["doc-lifecycle", "status-writeback", "notify", "custom"].map((k) => `<option value="${k}" ${a.kind === k ? "selected" : ""}>${k}</option>`).join("") + `</select><select class="inp" data-act-policy="${idx}" style="width:170px"><option value="invalidate" ${(a.action || cfg.bugDocPolicy) === "invalidate" ? "selected" : ""}>\u6807\u8BB0\u5931\u6548 (invalidate)</option><option value="delete" ${(a.action || cfg.bugDocPolicy) === "delete" ? "selected" : ""}>\u5220\u9664\u6587\u6863 (delete)</option><option value="keep" ${(a.action || cfg.bugDocPolicy) === "keep" ? "selected" : ""}>\u4FDD\u7559\u4E0D\u5220\u9664 (keep)</option></select><input class="inp" data-act-title="${idx}" value="${esc(a.title || "")}" placeholder="\u6536\u5C3E\u52A8\u4F5C\u8BF4\u660E" style="flex:2;min-width:220px" /><button type="button" class="btn tiny mute" data-act-del="${idx}">\u5220\u9664\u52A8\u4F5C</button></div>`
    ).join("");
    const designNodesHtml = (cfg.phases || []).map((n, i) => nodeRowHtml(n, i, cfg.provider)).join("");
    return `
  <h2 class="sec-title">${esc(t("\u8BBE\u8BA1\u8868\u4E0E\u8282\u70B9\u6D41\u8F6C\u67B6\u6784\uFF08Workflow + Scheme = \u58F0\u660E\u5C42\uFF0C\u4EE3\u7801 = \u901A\u7528\u89E3\u91CA\u5668\uFF09"))}</h2>
  <div class="card">
    <div class="card-h">
      <span class="h-title">${esc(t(String((cfg.phases || []).length) + " \u8282\u70B9\u76F8\u4F4D\u94FE\uFF08\u591A\u8D77\u70B9 / \u6761\u4EF6\u5206\u53C9\u8FB9\u2014\u2014\u5224\u65AD\u8BCD\u5F62=LLM \u5206\u652F\u8F93\u51FA\uFF1B\u70B9\u51FB\u8282\u70B9\u76D2\u6216\u8868\u884C\u5F39\u7A97\u7F16\u8F91\uFF09"))}</span>
      <button type="button" class="btn tiny acc" data-open-provider="${esc(cfg.provider)}">${esc(t("\u914D\u7F6E\u9ED8\u8BA4 OpenAIProvider (" + cfg.provider + ")"))}</button>
    </div>
    <div class="card-b" style="overflow-x:auto">
      <div class="wf-graph-scroll" data-node-graph>${wfGraphHtml(cfg.phases, cfg.transitions)}</div>
    </div>
  </div>

  <div class="card" style="margin-top:14px">
    <div class="card-h">
      <span class="h-title">${esc(t(String((cfg.phases || []).length) + " \u8282\u70B9\u4E3B\u5E72\u8BBE\u8BA1\u8868\uFF08\u70B9\u51FB\u4EFB\u610F\u884C\u5F39\u7A97\u7F16\u8F91\u8282\u70B9\u7EC6\u8282\u914D\u7F6E\uFF09"))}</span>
      <div style="display:flex;gap:8px;align-items:center">
        <button type="button" class="btn tiny acc" data-node-add>${esc(t("+ \u6DFB\u52A0\u8282\u70B9"))}</button>
        <button type="button" class="btn tiny acc" data-open-provider="${esc(cfg.provider)}">${esc(t("\u914D\u7F6E\u9ED8\u8BA4 OpenAIProvider (" + cfg.provider + ")"))}</button>
      </div>
    </div>
    <div class="card-b" style="overflow-x:auto">
      <div class="small faint" style="margin-bottom:8px">
        ${esc(t("\u5165\u53E3 1\uFF08\u8FD0\u884C\u671F\u65E5\u5FD7/Bug/\u5DE1\u68C0\u4E8B\u4EF6\uFF09\u4ECE\u8282\u70B9 1 (Idle) \u987A\u6B21\u6D41\u8F6C\uFF1B\u5165\u53E3 2\uFF08\u5BF9\u8BDD\u4FEE\u6539\u8BBE\u8BA1\u6587\u6863\u6216\u4EA7\u51FA\u4E34\u65F6 BugDoc \u89E6\u53D1 DocApplied\uFF09\u76F4\u63A5\u4ECE\u8282\u70B9 4 (Planning / Doc Apply) \u5207\u5165\uFF0C\u76F4\u8D70 developer-system \u2192 \u9A8C\u6536 \u2192 deploy-system \u2192 \u603B\u7ED3\u5F52\u6863\u3002\u70B9\u51FB\u4EFB\u610F Provider \u5FBD\u7AE0\u53EF\u76F4\u63A5\u914D\u7F6E OpenAIProvider \u8D26\u53F7\u3001Base \u5730\u5740\u4E0E\u6A21\u578B\u5206\u7EA7\u3002"))}
      </div>
      <table class="tbl">
        <thead><tr><th>\u8282\u70B9</th><th>\u76F8\u4F4D</th><th>\u804C\u8D23\u8BF4\u660E</th><th>\u6267\u884C\u5355\u5143</th><th>ConnectorOps / \u7B97\u5B50</th><th>OpenAIProvider / \u6A21\u578B\u5206\u7EA7</th><th>ref \u89C4\u5219 / ref tag</th></tr></thead>
        <tbody data-node-tbody>${designNodesHtml}</tbody>
      </table>
    </div>
  </div>

  <div data-provider-drawer></div>

  <h2 class="sec-title">${esc(t("\u4EE5\u4EBA\u4E3A\u672C\u53EF\u89C6\u5316\u914D\u7F6E\u5DE5\u4F5C\u53F0\uFF08Scheme Adapter \u5C42\u914D\u7F6E\uFF0C\u4FDD\u5B58\u5373 Apply \u843D\u5E93\uFF09"))}</h2>
  <div class="card" data-wf-studio>
    <div class="card-h">
      <span class="h-title">${esc(usedName)} \xB7 ${esc(t("\u5168\u8282\u70B9\u53C2\u6570\u3001\u81EA\u52A8\u5BA1\u6279\u3001\u9A8C\u6536\u6761\u4EF6\u4E0E\u4E34\u65F6 BugDoc \u6536\u5C3E\u7B56\u7565\u914D\u7F6E"))}</span>
      <div>
        ${cfg.undeletable ? chip(t("\u5185\u7F6E\u4FDD\u62A4\uFF1A\u4E0D\u53EF\u5220\u9664"), "warn") : ""}
        ${cfg.immutableNodes ? chip(t("\u4E3B\u5E72\u8282\u70B9\u62D3\u6251\u9501\u5B9A\uFF08\u53C2\u6570\u53EF\u914D\uFF09"), "acc") : chip(t("\u8282\u70B9\u53EF\u52A0\u5220\uFF08draft\u2192apply\uFF09"), "ok")}
      </div>
    </div>
    <div class="card-b">
      <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:12px;margin-bottom:14px">
        <div class="card" style="margin:0"><div class="card-b">
          <b>${esc(t("A. \u603B\u63A7\u5F00\u5173\u4E0E\u5168\u5C40\u89C4\u5219\u5F15\u7528"))}</b>
          <div style="margin-top:8px">
            <label style="display:flex;align-items:center;gap:8px;cursor:pointer">
              <input type="checkbox" data-cfg="enabled" ${cfg.enabled ? "checked" : ""} />
              <span><b>${esc(t("\u542F\u7528\u6B64\u5DE5\u4F5C\u6D41 (spec.enabled)"))}</b> ${esc(t("\uFF08\u51FA\u5382\u9ED8\u8BA4\u4E0D\u542F\u52A8\uFF1Bcore-system \u81EA\u6211\u8FED\u4EE3\u663E\u5F0F\u5F00\u542F\uFF09"))}</span>
            </label>
          </div>
          <div style="margin-top:8px">
            <div class="small faint">${esc(t("\u6D89\u53CA\u670D\u52A1 (services.involved)"))}</div>
            <input class="inp" data-cfg="servicesInvolved" value="${esc(cfg.servicesInvolved)}" style="width:100%" />
          </div>
          <div style="margin-top:8px">
            <div class="small faint">${esc(t("Session/Agent \u9ED8\u8BA4\u5F15\u7528\u5168\u5C40\u89C4\u5219 (globalRuleRefs)"))}</div>
            <input class="inp mono" data-cfg="globalRuleRefs" value="${esc(cfg.globalRuleRefs)}" style="width:100%" />
          </div>
        </div></div>

        <div class="card" style="margin:0"><div class="card-b">
          <b>${esc(t("B. \u8282\u70B9 1\uFF08\u76D1\u542C\u65E5\u5FD7\u4E0E Event \u8FC7\u6EE4 + ConnectorOps \u6536\u96C6\uFF09"))}</b>
          <div style="margin-top:8px;display:grid;grid-template-columns:1fr 110px;gap:6px">
            <div>
              <div class="small faint">${esc(t("\u76D1\u542C\u4E8B\u4EF6\u7C7B\u578B"))}</div>
              <input class="inp mono" data-cfg="logEvents" value="${esc(cfg.logEvents)}" style="width:100%" />
            </div>
            <div>
              <div class="small faint">${esc(t("\u6700\u4F4E\u65E5\u5FD7\u7EA7\u522B"))}</div>
              <select class="inp" data-cfg="logMinLevel" style="width:100%">
                ${["debug", "info", "warn", "error", "fatal"].map((l) => `<option value="${l}" ${cfg.logMinLevel === l ? "selected" : ""}>${l}</option>`).join("")}
              </select>
            </div>
          </div>
          <div style="margin-top:6px;display:grid;grid-template-columns:1fr 1fr;gap:6px">
            <div>
              <div class="small faint">${esc(t("\u8FC7\u6EE4 Service"))}</div>
              <input class="inp mono" data-cfg="logServices" value="${esc(cfg.logServices)}" style="width:100%" />
            </div>
            <div>
              <div class="small faint">${esc(t("\u65E5\u5FD7\u5173\u952E\u8BCD/\u6B63\u5219\u8FC7\u6EE4"))}</div>
              <input class="inp mono" data-cfg="logKeyword" value="${esc(cfg.logKeyword)}" style="width:100%" />
            </div>
          </div>
          <div style="margin-top:6px;display:grid;grid-template-columns:1fr 110px auto;gap:6px;align-items:end">
            <div>
              <div class="small faint">${esc(t("\u65E5\u5FD7\u6536\u96C6 ConnectorOps"))}</div>
              <input class="inp mono" data-cfg="collectorConnector" value="${esc(cfg.collectorConnector)}" style="width:100%" />
            </div>
            <div>
              <div class="small faint">${esc(t("\u6536\u96C6\u7B97\u5B50 op"))}</div>
              <input class="inp mono" data-cfg="collectorOp" value="${esc(cfg.collectorOp)}" style="width:100%" />
            </div>
            <button type="button" class="btn tiny acc" data-develop-collector title="\u5F53\u65E5\u5FD7\u6536\u96C6 ConnectorOps \u7F3A\u5931\u6216\u4E3A\u9AA8\u67B6\u65F6\uFF0C\u76F4\u63A5\u62C9\u51FA\u4EE3\u7801\u5F00\u53D1\u4EFB\u52A1\u5E76\u6253\u5F00\u7F16\u8F91">${esc(t("\u62C9\u51FA\u4EE3\u7801\u5F00\u53D1/\u7F16\u8F91"))}</button>
          </div>
          <div style="margin-top:6px;display:grid;grid-template-columns:auto 100px 1fr;gap:6px;align-items:center">
            <label><input type="checkbox" data-cfg="timerEnabled" ${cfg.timerEnabled ? "checked" : ""} /> ${esc(t("\u5B9A\u65F6\u89E6\u53D1"))}</label>
            <input class="inp mono" data-cfg="timerInterval" value="${esc(cfg.timerInterval)}" placeholder="30m" />
            <input class="inp mono" data-cfg="timerCron" value="${esc(cfg.timerCron)}" placeholder="*/30 * * * *" />
          </div>
        </div></div>

        <div class="card" style="margin:0"><div class="card-b">
          <b>${esc(t("C. \u8282\u70B9 3\uFF08\u5BA1\u6279 / \u81EA\u52A8\u5224\u65AD\u5BA1\u6279\uFF09& \u8282\u70B9 4\uFF08Doc Apply \u7B2C\u4E8C\u5165\u53E3\uFF09"))}</b>
          <div style="margin-top:8px;display:grid;grid-template-columns:1fr 110px;gap:6px">
            <div>
              <div class="small faint">${esc(t("\u8282\u70B9 3 \u5BA1\u6279\u6A21\u5F0F (approval.mode)"))}</div>
              <select class="inp" data-cfg="approvalMode" style="width:100%">
                <option value="agent" ${cfg.approvalMode === "agent" ? "selected" : ""}>agent\uFF08Agent \u6309\u89C4\u5219\u81EA\u52A8\u5224\u65AD\u5BA1\u6279\uFF09</option>
                <option value="auto" ${cfg.approvalMode === "auto" ? "selected" : ""}>auto\uFF08\u5168\u81EA\u52A8\u653E\u884C\u4E0D\u6302\u8D77\uFF09</option>
                <option value="human" ${cfg.approvalMode === "human" ? "selected" : ""}>human\uFF08\u4EBA\u5DE5\u5BA1\u6279\u6302\u8D77\uFF09</option>
              </select>
            </div>
            <div>
              <div class="small faint">${esc(t("\u81EA\u52A8\u653E\u884C\u6700\u9AD8\u98CE\u9669"))}</div>
              <select class="inp" data-cfg="approvalMaxRisk" style="width:100%">
                ${["low", "medium", "high"].map((r) => `<option value="${r}" ${cfg.approvalMaxRisk === r ? "selected" : ""}>${r}</option>`).join("")}
              </select>
            </div>
          </div>
          <div style="margin-top:6px">
            <div class="small faint">${esc(t("Agent \u81EA\u52A8\u5BA1\u6279\u5224\u65AD\u51C6\u5219 (autoCriteria)"))}</div>
            <input class="inp" data-cfg="approvalCriteria" value="${esc(cfg.approvalCriteria)}" style="width:100%" />
          </div>
          <div style="margin-top:8px;padding-top:8px;border-top:1px solid var(--ds-border)">
            <label style="display:flex;align-items:center;gap:6px;cursor:pointer">
              <input type="checkbox" data-cfg="docApplyEntryEnabled" ${cfg.docApplyEntryEnabled ? "checked" : ""} />
              <span><b>${esc(t("\u542F\u7528\u8282\u70B9 4 (Doc Apply) \u591A\u5165\u53E3\u5207\u5165"))}</b> ${esc(t("\uFF08\u5BF9\u8BDD\u6539\u6587\u6863/\u4E34\u65F6 BugDoc \u89E6\u53D1 DocApplied \u76F4\u5207\u8282\u70B9 4\uFF09"))}</span>
            </label>
            <input class="inp mono" data-cfg="docApplyDocTypes" value="${esc(cfg.docApplyDocTypes)}" style="width:100%;margin-top:4px" />
          </div>
        </div></div>

        <div class="card" style="margin:0"><div class="card-b">
          <b>${esc(t("D. \u8282\u70B9 5 & 7\uFF08developer-system / deploy-system \u9010\u5C42\u7ED1\u5B9A\uFF09"))}</b>
          <div style="margin-top:8px;display:grid;grid-template-columns:1fr 130px;gap:6px">
            <div>
              <div class="small faint">${esc(t("\u7814\u53D1\u7CFB\u7EDF (developer-system)"))}</div>
              <input class="inp mono" data-cfg="devService" value="${esc(cfg.devService)}" style="width:100%" />
            </div>
            <div>
              <div class="small faint">${esc(t("\u6765\u6E90\u7B56\u7565 (\u9010\u5C42\u8BE2\u95EE)"))}</div>
              <select class="inp" data-cfg="devResolution" style="width:100%">
                <option value="default" ${cfg.devResolution === "default" ? "selected" : ""}>default (\u9ED8\u8BA4\u5F00\u6E90\u4E60\u60EF)</option>
                <option value="install" ${cfg.devResolution === "install" ? "selected" : ""}>install (\u6A21\u677F\u5B89\u88C5)</option>
                <option value="import" ${cfg.devResolution === "import" ? "selected" : ""}>import (\u5BFC\u5165\u5DF2\u6709\u7CFB\u7EDF)</option>
              </select>
            </div>
          </div>
          <div style="margin-top:6px;display:grid;grid-template-columns:1fr 1fr 80px;gap:6px">
            <input class="inp mono" data-cfg="devBranchPrefix" value="${esc(cfg.devBranchPrefix)}" placeholder="\u5206\u652F\u524D\u7F00 agent/" title="\u5F00\u6E90\u4E60\u60EF\uFF1A\u5206\u652F\u524D\u7F00" />
            <input class="inp mono" data-cfg="devWorktreeDir" value="${esc(cfg.devWorktreeDir)}" placeholder="worktrees/" title="\u5F00\u6E90\u4E60\u60EF\uFF1Aworktree \u76EE\u5F55" />
            <input class="inp mono" data-cfg="devMaxLines" value="${esc(String(cfg.devMaxLines))}" placeholder="800" title="\u5355\u6587\u4EF6\u884C\u6570\u4E0A\u9650" />
          </div>
          <div style="margin-top:8px;display:grid;grid-template-columns:1fr 130px;gap:6px">
            <div>
              <div class="small faint">${esc(t("\u53D1\u5E03\u7CFB\u7EDF (deploy-system)"))}</div>
              <input class="inp mono" data-cfg="deployService" value="${esc(cfg.deployService)}" style="width:100%" />
            </div>
            <div>
              <div class="small faint">${esc(t("\u6765\u6E90\u7B56\u7565 (\u9010\u5C42\u8BE2\u95EE)"))}</div>
              <select class="inp" data-cfg="deployResolution" style="width:100%">
                <option value="default" ${cfg.deployResolution === "default" ? "selected" : ""}>default (\u9ED8\u8BA4\u5BBF\u4E3B\u53D1\u5E03)</option>
                <option value="install" ${cfg.deployResolution === "install" ? "selected" : ""}>install (\u6A21\u677F\u5B89\u88C5)</option>
                <option value="import" ${cfg.deployResolution === "import" ? "selected" : ""}>import (\u5BFC\u5165\u5916\u90E8\u53D1\u5E03)</option>
              </select>
            </div>
          </div>
        </div></div>
      </div>

      <div class="card" style="margin-bottom:12px"><div class="card-b">
        <div style="display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:8px;margin-bottom:8px">
          <b>${esc(t("E. \u8282\u70B9 6\uFF08\u9A8C\u6536 Verify \u914D\u7F6E \u2014\u2014 \u600E\u4E48\u9A8C\u6536\u3001\u9A8C\u6536\u6761\u4EF6\u53EF\u81EA\u7531\u6DFB\u52A0/\u5220\u9664\uFF09"))}</b>
          <div style="display:flex;gap:8px;align-items:center">
            <select class="inp" data-cfg="verifyMode">
              <option value="both" ${cfg.verifyMode === "both" ? "selected" : ""}>both\uFF08\u4EE3\u7801\u5BA1\u6838 + Agent \u9A8C\u6536\uFF09</option>
              <option value="code" ${cfg.verifyMode === "code" ? "selected" : ""}>code\uFF08\u4EC5\u4EE3\u7801\u89C4\u5219\u5BA1\u6838\uFF09</option>
              <option value="agent" ${cfg.verifyMode === "agent" ? "selected" : ""}>agent\uFF08\u4EC5 Agent \u63A2\u9488\u9A8C\u6536\uFF09</option>
              <option value="human" ${cfg.verifyMode === "human" ? "selected" : ""}>human\uFF08\u4EBA\u5DE5\u786E\u8BA4\u9A8C\u6536\uFF09</option>
            </select>
            <select class="inp" data-cfg="verifyOnFail">
              <option value="retry-develop" ${cfg.verifyOnFail === "retry-develop" ? "selected" : ""}>\u672A\u901A\u8FC7\uFF1A\u6253\u56DE developer-system \u91CD\u4FEE</option>
              <option value="block" ${cfg.verifyOnFail === "block" ? "selected" : ""}>\u672A\u901A\u8FC7\uFF1A\u963B\u65AD\u5E76\u4E0A\u62A5</option>
              <option value="ask" ${cfg.verifyOnFail === "ask" ? "selected" : ""}>\u672A\u901A\u8FC7\uFF1A\u8F6C\u4EBA\u5DE5\u88C1\u51B3</option>
            </select>
            <button type="button" class="btn tiny acc" data-cond-add>${esc(t("+ \u6DFB\u52A0\u9A8C\u6536\u6761\u4EF6"))}</button>
          </div>
        </div>
        <div data-wf-cond-list>${condRowsHtml}</div>
      </div></div>

      <div class="card" style="margin-bottom:12px"><div class="card-b">
        <div style="display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:8px;margin-bottom:8px">
          <b>${esc(t("F. \u8282\u70B9 8\uFF08\u603B\u7ED3\u5F52\u6863 Reporting \u901A\u7528\u89E3\u91CA\u5668\u6536\u5C3E\u52A8\u4F5C \u2014\u2014 \u4E34\u65F6 BugDoc \u9ED8\u8BA4\u53EF\u6807\u8BB0\u5931\u6548/\u5220\u9664/\u9009\u62E9\u4E0D\u5220\u9664\uFF0C\u52A8\u4F5C\u9879\u53EF\u81EA\u7531\u589E\u5220\uFF09"))}</b>
          <div style="display:flex;gap:8px;align-items:center">
            <span class="small faint">${esc(t("\u4E34\u65F6 BugDoc \u9ED8\u8BA4\u7B56\u7565\uFF1A"))}</span>
            <select class="inp" data-cfg="bugDocPolicy">
              <option value="invalidate" ${cfg.bugDocPolicy === "invalidate" ? "selected" : ""}>\u6807\u8BB0\u5931\u6548\u5F52\u6863 (invalidate)</option>
              <option value="delete" ${cfg.bugDocPolicy === "delete" ? "selected" : ""}>\u5220\u9664\u4E34\u65F6 BugDoc (delete)</option>
              <option value="keep" ${cfg.bugDocPolicy === "keep" ? "selected" : ""}>\u4FDD\u7559\u4E0D\u5220\u9664 (keep)</option>
            </select>
            <label style="display:flex;align-items:center;gap:4px"><input type="checkbox" data-cfg="statusWriteback" ${cfg.statusWriteback ? "checked" : ""} /> ${esc(t("\u56DE\u5199 InSync"))}</label>
            <button type="button" class="btn tiny acc" data-act-add>${esc(t("+ \u6DFB\u52A0\u6536\u5C3E\u52A8\u4F5C"))}</button>
          </div>
        </div>
        <div data-wf-act-list>${postActRowsHtml}</div>
      </div></div>

      <details class="ex-block" open>
        <summary class="ex-sum"><span class="ex-chev">\u25B6</span><b>${esc(t("Workflow \u5B8C\u6574 YAML \u58F0\u660E\uFF08\u89E3\u91CA\u5668\u6743\u5A01\u6E90\uFF0C\u8868\u5355\u8054\u52A8\u5B9E\u65F6\u751F\u6210\uFF0C\u4E5F\u53EF\u76F4\u63A5\u7F16\u8F91 YAML\uFF09"))}</b></summary>
        <div class="ex-body">
          <textarea class="inp mono" data-wf-yaml rows="18" style="width:100%;font-size:12px;line-height:1.45">${esc(yamlText)}</textarea>
        </div>
      </details>

      <div style="margin-top:12px;display:flex;gap:10px;align-items:center;flex-wrap:wrap">
        <button type="button" class="btn acc" data-wf-save>${esc(t("\u4FDD\u5B58\u5E76 Apply \u5DE5\u4F5C\u6D41\u914D\u7F6E"))}</button>
        <button type="button" class="btn mute" data-wf-sync-yaml>${esc(t("\u4ECE\u4E0A\u65B9\u8868\u5355\u91CD\u65B0\u751F\u6210 YAML"))}</button>
        <span class="small faint" data-wf-save-msg></span>
      </div>
    </div>
  </div>`;
  }
  function bindWorkflowStudio(el, ctx2, wf, usedName, onSaved) {
    const studio = el.querySelector("[data-wf-studio]");
    if (!studio) return;
    const scope = wf && wf.primaryScope || "global";
    let baseRv = Number(wf && wf.primaryRv || 0);
    const cfg = defaultLoopSpec(usedName, wf && wf.spec);
    const readFormIntoCfg = () => {
      studio.querySelectorAll("[data-cfg]").forEach((inp) => {
        const k = inp.getAttribute("data-cfg");
        if (inp.type === "checkbox") cfg[k] = Boolean(inp.checked);
        else cfg[k] = inp.value;
      });
      const conds = [];
      studio.querySelectorAll("[data-wf-cond-row]").forEach((row) => {
        const idx = row.getAttribute("data-wf-cond-row");
        conds.push({
          id: "cond-" + (conds.length + 1),
          required: Boolean(row.querySelector(`[data-cond-req="${idx}"]`)?.checked),
          kind: row.querySelector(`[data-cond-kind="${idx}"]`)?.value || "code",
          title: row.querySelector(`[data-cond-title="${idx}"]`)?.value || "",
          command: row.querySelector(`[data-cond-cmd="${idx}"]`)?.value || ""
        });
      });
      cfg.conditions = conds;
      const acts = [];
      studio.querySelectorAll("[data-wf-act-row]").forEach((row) => {
        const idx = row.getAttribute("data-wf-act-row");
        acts.push({
          id: "act-" + (acts.length + 1),
          enabled: Boolean(row.querySelector(`[data-act-en="${idx}"]`)?.checked),
          kind: row.querySelector(`[data-act-kind="${idx}"]`)?.value || "doc-lifecycle",
          targetDocType: "BugDoc",
          action: row.querySelector(`[data-act-policy="${idx}"]`)?.value || cfg.bugDocPolicy || "invalidate",
          title: row.querySelector(`[data-act-title="${idx}"]`)?.value || ""
        });
      });
      cfg.postActions = acts;
    };
    const syncYamlBox = () => {
      readFormIntoCfg();
      const box = studio.querySelector("[data-wf-yaml]");
      if (box) box.value = buildWorkflowYaml(cfg, scope);
    };
    studio.querySelectorAll("[data-cfg]").forEach((inp) => {
      inp.addEventListener("change", syncYamlBox);
    });
    const bindRowDeletes = () => {
      studio.querySelectorAll("[data-cond-del]").forEach((btn) => {
        btn.onclick = () => {
          btn.closest("[data-wf-cond-row]")?.remove();
          syncYamlBox();
        };
      });
      studio.querySelectorAll("[data-act-del]").forEach((btn) => {
        btn.onclick = () => {
          btn.closest("[data-wf-act-row]")?.remove();
          syncYamlBox();
        };
      });
      studio.querySelectorAll("[data-wf-cond-row] input, [data-wf-cond-row] select, [data-wf-act-row] input, [data-wf-act-row] select").forEach((inp) => {
        inp.onchange = syncYamlBox;
      });
    };
    bindRowDeletes();
    studio.querySelector("[data-cond-add]")?.addEventListener("click", () => {
      const list2 = studio.querySelector("[data-wf-cond-list]");
      if (!list2) return;
      const idx = Date.now();
      const div = document.createElement("div");
      div.className = "mini-row";
      div.setAttribute("data-wf-cond-row", String(idx));
      div.style.cssText = "display:flex;gap:8px;align-items:center;flex-wrap:wrap";
      div.innerHTML = `<input type="checkbox" data-cond-req="${idx}" checked /><select class="inp" data-cond-kind="${idx}" style="width:110px"><option value="code">code</option><option value="test">test</option><option value="agent">agent</option><option value="probe">probe</option><option value="custom">custom</option></select><input class="inp" data-cond-title="${idx}" value="\u81EA\u5B9A\u4E49\u9A8C\u6536\u6761\u4EF6" style="flex:2;min-width:200px" /><input class="inp mono" data-cond-cmd="${idx}" value="connectops-code-audit/audit" style="flex:1;min-width:160px" /><button type="button" class="btn tiny mute" data-cond-del="${idx}">\u5220\u9664</button>`;
      list2.appendChild(div);
      bindRowDeletes();
      syncYamlBox();
    });
    studio.querySelector("[data-act-add]")?.addEventListener("click", () => {
      const list2 = studio.querySelector("[data-wf-act-list]");
      if (!list2) return;
      const idx = Date.now();
      const div = document.createElement("div");
      div.className = "mini-row";
      div.setAttribute("data-wf-act-row", String(idx));
      div.style.cssText = "display:flex;gap:8px;align-items:center;flex-wrap:wrap";
      div.innerHTML = `<input type="checkbox" data-act-en="${idx}" checked /><select class="inp" data-act-kind="${idx}" style="width:140px"><option value="doc-lifecycle">doc-lifecycle</option><option value="status-writeback">status-writeback</option><option value="notify">notify</option><option value="custom">custom</option></select><select class="inp" data-act-policy="${idx}" style="width:170px"><option value="invalidate">\u6807\u8BB0\u5931\u6548 (invalidate)</option><option value="delete">\u5220\u9664\u6587\u6863 (delete)</option><option value="keep">\u4FDD\u7559\u4E0D\u5220\u9664 (keep)</option></select><input class="inp" data-act-title="${idx}" value="\u4E34\u65F6 BugDoc \u6536\u5C3E\u52A8\u4F5C" style="flex:2;min-width:220px" /><button type="button" class="btn tiny mute" data-act-del="${idx}">\u5220\u9664\u52A8\u4F5C</button>`;
      list2.appendChild(div);
      bindRowDeletes();
      syncYamlBox();
    });
    studio.querySelector("[data-wf-sync-yaml]")?.addEventListener("click", syncYamlBox);
    const repaintNodeSurfaces = () => {
      const tbl = el.querySelector("[data-node-tbody]");
      if (tbl) tbl.innerHTML = (cfg.phases || []).map((n, i) => nodeRowHtml(n, i, cfg.provider)).join("");
      const graph = el.querySelector("[data-node-graph]");
      if (graph) graph.innerHTML = wfGraphHtml(cfg.phases, cfg.transitions);
      bindRowClicks();
    };
    const closeModal = () => {
      el.querySelectorAll(".chat-modal[data-nm-modal]").forEach((m) => m.remove());
    };
    const openNodeModal = (nameOrIdx) => {
      readFormIntoCfg();
      const idx = typeof nameOrIdx === "number" ? nameOrIdx : cfg.phases.findIndex((n) => n.name === nameOrIdx);
      if (idx < 0 || !cfg.phases[idx]) return;
      closeModal();
      const modal = document.createElement("div");
      modal.className = "chat-modal";
      modal.setAttribute("data-nm-modal", "1");
      const paint3 = (node) => {
        modal.innerHTML = '<div class="chat-modal-mask"></div><div class="chat-modal-panel"><div class="chat-modal-head"><div class="chat-modal-title"><span class="chip ' + (UNIT_CHIP[node.unit] || "mute") + ' tiny">' + esc(node.unit === "gate" ? t("\u63A7\u5236\u8282\u70B9") : node.unit || "connectops") + '</span> <b class="mono">' + esc(node.name || "") + "</b> \xB7 " + esc(t("\u8282\u70B9\u7EC6\u8282\u914D\u7F6E")) + '</div><div class="chat-modal-acts"><button type="button" class="btn tiny mute" data-nm-cancel>' + esc(t("\u53D6\u6D88")) + '</button></div></div><div class="chat-modal-body"><div data-nm-fields>' + nodeFieldsHtml(node, cfg.provider) + '</div><div style="margin-top:14px;display:flex;gap:8px;align-items:center"><button type="button" class="btn acc sm" data-nm-save>' + esc(t("\u4FDD\u5B58\u8282\u70B9")) + '</button><button type="button" class="btn btn-ghost btn-danger sm" data-nm-del>' + esc(t("\u5220\u9664\u8282\u70B9")) + '</button><span class="small faint">' + esc(t("\u6539\u52A8\u968F\u300C\u4FDD\u5B58\u5E76 Apply \u5DE5\u4F5C\u6D41\u914D\u7F6E\u300D\u4EE5 Workflow \u6587\u6863\u5BF9\u8C61\u843D\u8D26")) + "</span></div></div></div>";
        modal.querySelector("[data-nm-cancel]").addEventListener("click", closeModal);
        modal.querySelector(".chat-modal-mask").addEventListener("click", closeModal);
        modal.querySelector('[data-nm="unit"]').addEventListener("change", () => {
          const draft = readNodeFields(modal, cfg.phases[idx]);
          cfg.phases[idx] = draft;
          paint3(cfg.phases[idx]);
        });
        modal.querySelector("[data-nm-save]").addEventListener("click", () => {
          cfg.phases[idx] = readNodeFields(modal, cfg.phases[idx]);
          closeModal();
          repaintNodeSurfaces();
          syncYamlBox();
        });
        modal.querySelector("[data-nm-del]").addEventListener("click", () => {
          cfg.phases.splice(idx, 1);
          closeModal();
          repaintNodeSurfaces();
          syncYamlBox();
        });
      };
      paint3(cfg.phases[idx]);
      el.appendChild(modal);
    };
    const bindRowClicks = () => {
      el.querySelectorAll("[data-node-edit]").forEach((row) => {
        row.onclick = () => openNodeModal(row.getAttribute("data-node-edit") || "");
      });
      el.querySelectorAll("[data-graph-node]").forEach((box) => {
        box.onclick = () => openNodeModal(box.getAttribute("data-graph-node") || "");
      });
    };
    el.querySelector("[data-node-add]")?.addEventListener("click", () => {
      readFormIntoCfg();
      cfg.phases.push({ name: "node-" + (cfg.phases.length + 1), phase: "Idle", title: "\u65B0\u8282\u70B9", unit: "connectops", provider: cfg.provider, modelTier: "standard" });
      repaintNodeSurfaces();
      syncYamlBox();
      openNodeModal(cfg.phases.length - 1);
    });
    bindRowClicks();
    studio.querySelector("[data-develop-collector]")?.addEventListener("click", async () => {
      readFormIntoCfg();
      const connName = String(cfg.collectorConnector || "connectops-log-collector").trim();
      const opName = String(cfg.collectorOp || "collect-logs").trim();
      const msg = studio.querySelector("[data-wf-save-msg]");
      if (msg) msg.textContent = t("\u6B63\u5728\u68C0\u67E5/\u751F\u6210 ConnectorOps \u9AA8\u67B6\u5E76\u62C9\u51FA\u4EE3\u7801\u5F00\u53D1\u4EFB\u52A1\u2026");
      const existing = await ctx2.api.object("ConnectorOps", "global", connName);
      if (!existing.ok) {
        const connYaml = `apiVersion: core/v1
kind: ConnectorOps
metadata:
  name: ${connName}
  scope: global
spec:
  workservice: workflow-engine
  workflow: ${usedName}
  provider: ${cfg.provider || "openai-default"}
  llm:
    provider: ${cfg.provider || "openai-default"}
    tier: low
    model: "flash-tier"
  ops: [${opName}]
  config:
    query: ${JSON.stringify(cfg.collectorQuery || "level>=warn")}
    services: ${JSON.stringify(cfg.logServices || "self-recursive-core")}
`;
        await ctx2.api.apply({ kind: "ConnectorOps", scope: "global", name: connName, resourceVersion: 0, content: connYaml, fieldManager: "console:workflow-studio", reasonRef: "develop-collector:" + connName });
      }
      if (msg) msg.innerHTML = esc(t("\u5DF2\u5C31\u7EEA ConnectorOps/global/" + connName + " \u2014\u2014 ")) + chipGo(t("\u6253\u5F00 ConnectorOps \u7F16\u8F91\u5668\u5F00\u53D1\u7B97\u5B50"), docRoute("ConnectorOps", "global", connName), "acc", true);
    });
    studio.querySelector("[data-wf-save]")?.addEventListener("click", async () => {
      const yamlBox = studio.querySelector("[data-wf-yaml]");
      const content = yamlBox ? yamlBox.value : buildWorkflowYaml(cfg, scope);
      const msg = studio.querySelector("[data-wf-save-msg]");
      if (msg) msg.textContent = t("\u6B63\u5728 Apply \u5DE5\u4F5C\u6D41\u914D\u7F6E\u2026");
      let rv = baseRv;
      const cur = await ctx2.api.object("Workflow", scope, usedName);
      if (cur.ok && cur.data && cur.data.resourceVersion != null) rv = Number(cur.data.resourceVersion) || 0;
      const res = await ctx2.api.apply({
        kind: "Workflow",
        scope,
        name: usedName,
        resourceVersion: rv,
        content,
        fieldManager: "console:workflow-studio",
        reasonRef: "workflow-config:" + usedName
      });
      if (!res.ok) {
        if (msg) msg.textContent = t("Apply \u5931\u8D25\uFF1A") + String(res.error || res.status);
        return;
      }
      baseRv = Number(res.data && res.data.rv) || rv + 1;
      if (msg) msg.textContent = t("\u5DF2\u4FDD\u5B58\u5E76 Apply \u751F\u6548\uFF08rv=" + baseRv + "\uFF09");
      if (typeof onSaved === "function") onSaved();
    });
    el.querySelectorAll("[data-open-provider]").forEach((btn) => {
      btn.addEventListener("click", () => openProviderDrawer(el, ctx2, btn.getAttribute("data-open-provider") || "openai-default"));
    });
    el.querySelectorAll("[data-open-conn]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const cName = btn.getAttribute("data-open-conn");
        if (cName && ctx2.router) ctx2.router.go(docRoute("ConnectorOps", "global", cName));
      });
    });
  }
  async function openProviderDrawer(el, ctx2, providerName = "openai-default") {
    const slot = el.querySelector("[data-provider-drawer]");
    if (!slot) return;
    slot.innerHTML = `<div class="card" style="border:1px solid var(--ds-accent)"><div class="card-b"><div class="small faint">${esc(t("\u6B63\u5728\u52A0\u8F7D OpenAIProvider (" + providerName + ") \u914D\u7F6E\u2026"))}</div></div></div>`;
    const res = await ctx2.api.object("Provider", "global", providerName);
    const rv = res.ok && res.data && res.data.resourceVersion != null ? Number(res.data.resourceVersion) : 0;
    const raw = res.ok && res.data && res.data.content ? String(res.data.content) : "";
    const extract = (re, def2) => {
      const m = raw.match(re);
      return m ? m[1].trim().replace(/^["']|["']$/g, "") : def2;
    };
    const account = extract(/account:\s*([^\n#]+)/, "magpie-openai-gateway");
    const secretSrc = extract(/source:\s*([^\n}#]+)/, "secret:llm/magpie-gateway-key");
    const baseUrl = extract(/baseUrl:\s*([^\n#]+)/, "https://magpie.ListenAll.cn/v1");
    const lowModel = extract(/low:\s*([^\n}#]+)/, "flash-tier");
    const stdModel = extract(/standard:\s*([^\n}#]+)/, "standard-tier");
    const highModel = extract(/high:\s*([^\n}#]+)/, "reasoning-tier");
    slot.innerHTML = `
  <div class="card" style="border:1px solid var(--ds-accent);margin-bottom:14px">
    <div class="card-h">
      <span class="h-title">${esc(t("OpenAIProvider \u914D\u7F6E\u9762\u677F"))} \xB7 <span class="mono">${esc(providerName)}</span> (rv=${rv})</span>
      <button type="button" class="btn tiny mute" data-prov-close>${esc(t("\u6536\u8D77"))}</button>
    </div>
    <div class="card-b">
      <div class="small faint" style="margin-bottom:8px">${esc(t("\u6240\u6709 Session\u3001ConnectorOps \u4E0E Workflow \u8282\u70B9\u9ED8\u8BA4\u7ED1\u5B9A\u6B64 OpenAIProvider\u3002\u652F\u6301\u6309\u4EFB\u52A1\u667A\u5546\u8981\u6C42\u5206\u7EA7\u8DEF\u7531\uFF1A\u91CD\u590D\u6027/\u65E5\u5FD7\u8FC7\u6EE4\u4EFB\u52A1\u7528\u4F4E\u667A\u5546\u8F7B\u91CF\u6A21\u578B\uFF08low\uFF09\uFF0C\u5E38\u89C4\u5F00\u53D1\u7528\u6807\u51C6\u6A21\u578B\uFF08standard\uFF09\uFF0C\u67B6\u6784\u8BBE\u8BA1\u4E0E\u590D\u6742\u6839\u56E0\u8BCA\u65AD\u7528\u9AD8\u667A\u5546\u63A8\u7406\u6A21\u578B\uFF08high\uFF09\u3002"))}</div>
      <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:10px">
        <div>
          <div class="small faint">${esc(t("\u8D26\u53F7\u6807\u8BC6 (auth.account)"))}</div>
          <input class="inp mono" data-prov="account" value="${esc(account)}" style="width:100%" />
        </div>
        <div>
          <div class="small faint">${esc(t("\u5BC6\u94A5\u5F15\u7528 (auth.secret.source)"))}</div>
          <input class="inp mono" data-prov="secretSrc" value="${esc(secretSrc)}" style="width:100%" />
        </div>
        <div>
          <div class="small faint">${esc(t("OpenAI \u517C\u5BB9 Base \u5730\u5740 (config.baseUrl)"))}</div>
          <input class="inp mono" data-prov="baseUrl" value="${esc(baseUrl)}" style="width:100%" />
        </div>
      </div>
      <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:10px;margin-top:10px">
        <div>
          <div class="small faint">${esc(t("Low \u6863\u6A21\u578B\uFF08\u91CD\u590D\u6027\u4EFB\u52A1/\u65E5\u5FD7\u6536\u96C6/\u4EE3\u7801\u5BA1\u6838\u7B97\u5B50\uFF09"))}</div>
          <input class="inp mono" data-prov="lowModel" value="${esc(lowModel)}" style="width:100%" />
        </div>
        <div>
          <div class="small faint">${esc(t("Standard \u6863\u6A21\u578B\uFF08Doc Apply/\u5E38\u89C4\u5F00\u53D1\u7B97\u5B50\uFF09"))}</div>
          <input class="inp mono" data-prov="stdModel" value="${esc(stdModel)}" style="width:100%" />
        </div>
        <div>
          <div class="small faint">${esc(t("High \u6863\u6A21\u578B\uFF08\u67B6\u6784\u8BBE\u8BA1/Agent \u5224\u65AD/\u81EA\u52A8\u5BA1\u6279/\u6838\u5FC3\u7814\u53D1\uFF09"))}</div>
          <input class="inp mono" data-prov="highModel" value="${esc(highModel)}" style="width:100%" />
        </div>
      </div>
      <div style="margin-top:10px;display:flex;gap:10px;align-items:center">
        <button type="button" class="btn acc" data-prov-save>${esc(t("\u4FDD\u5B58\u5E76 Apply OpenAIProvider"))}</button>
        ${chipGo(t("\u5728\u5BF9\u8C61\u7F16\u8F91\u5668\u4E2D\u67E5\u770B YAML"), docRoute("Provider", "global", providerName), "mute", true)}
        <span class="small faint" data-prov-msg></span>
      </div>
    </div>
  </div>`;
    slot.querySelector("[data-prov-close]")?.addEventListener("click", () => {
      slot.innerHTML = "";
    });
    slot.querySelector("[data-prov-save]")?.addEventListener("click", async () => {
      const val = (k) => slot.querySelector(`[data-prov="${k}"]`)?.value.trim() || "";
      const yamlContent = `apiVersion: core/v1
kind: Provider
metadata:
  name: ${providerName}
  scope: global
spec:
  kind: openai
  trustLevel: trusted
  auth:
    scheme: bearer
    account: ${JSON.stringify(val("account") || "magpie-openai-gateway")}
    secret:
      source: ${JSON.stringify(val("secretSrc") || "secret:llm/magpie-gateway-key")}
      rotation: manual
  config:
    baseUrl: ${JSON.stringify(val("baseUrl") || "https://magpie.ListenAll.cn/v1")}
    defaultModel: ${JSON.stringify(val("highModel") || "reasoning-tier")}
    models: [${JSON.stringify(val("lowModel"))}, ${JSON.stringify(val("stdModel"))}, ${JSON.stringify(val("highModel"))}]
    modelTiers:
      low: ${JSON.stringify(val("lowModel") || "flash-tier")}
      standard: ${JSON.stringify(val("stdModel") || "standard-tier")}
      high: ${JSON.stringify(val("highModel") || "reasoning-tier")}
    taskModelMapping:
      connectops-log-collector: ${JSON.stringify(val("lowModel") || "flash-tier")}
      connectops-code-audit: ${JSON.stringify(val("lowModel") || "flash-tier")}
      connectops-deploy: ${JSON.stringify(val("lowModel") || "flash-tier")}
      doc-apply: ${JSON.stringify(val("stdModel") || "standard-tier")}
      agent-judge: ${JSON.stringify(val("highModel") || "reasoning-tier")}
      approval-gate: ${JSON.stringify(val("highModel") || "reasoning-tier")}
      connectops-developer: ${JSON.stringify(val("highModel") || "reasoning-tier")}
`;
      const msg = slot.querySelector("[data-prov-msg]");
      if (msg) msg.textContent = t("\u6B63\u5728 Apply OpenAIProvider\u2026");
      const r = await ctx2.api.apply({
        kind: "Provider",
        scope: "global",
        name: providerName,
        resourceVersion: rv,
        content: yamlContent,
        fieldManager: "console:provider-drawer",
        reasonRef: "openai-provider:" + providerName
      });
      if (msg) msg.textContent = r.ok ? t("OpenAIProvider \u5DF2\u4FDD\u5B58\u751F\u6548 (rv=" + (r.data && r.data.rv) + ")") : t("\u4FDD\u5B58\u5931\u8D25\uFF1A") + String(r.error || r.status);
    });
  }

  // src/screens/chat-nodes.js
  function blocksOf(content) {
    const out = { text: "", reasoning: "", toolCalls: [] };
    if (!Array.isArray(content)) return out;
    for (const b of content) {
      if (!b || typeof b !== "object") continue;
      if (b.type === "text" && typeof b.text === "string") out.text += (out.text ? "\n" : "") + b.text;
      else if (b.type === "reasoning" && typeof b.text === "string") out.reasoning += (out.reasoning ? "\n" : "") + b.text;
      else if (b.type === "tool-call") out.toolCalls.push({ id: String(b.id || b.callId || ""), name: String(b.name || ""), argsRaw: typeof b.arguments === "string" ? b.arguments : JSON.stringify(b.arguments ?? b.args ?? "") });
    }
    return out;
  }
  function nodeKey(n) {
    return String(Number(n && n.seq) || 0) + ":" + String(n && n.kind || "");
  }
  function resultTextOf(data) {
    const d = data && typeof data === "object" ? data : {};
    if (typeof d.content === "string" && d.content) return d.content;
    const msg = d.message && typeof d.message === "object" ? d.message : null;
    if (msg && Array.isArray(msg.content)) {
      const text2 = msg.content.filter((b) => b && b.type === "text" && typeof b.text === "string").map((b) => b.text).join("\n");
      if (text2) return text2;
    }
    if (d.output != null) return typeof d.output === "string" ? d.output : JSON.stringify(d.output);
    return "";
  }
  function triggerOf(source) {
    const s = source && typeof source === "object" ? source : {};
    const kind = String(s.kind || "");
    if (kind === "tool-jobs") return { id: "job" };
    if (kind === "cordis-host-runner") return { id: "plugin" };
    if (kind === "schedule") return { id: "schedule" };
    if (kind === "subagent-settled") return { id: "subagent", stopReason: String(s.stopReason || "") };
    if (kind === "goal") return { id: "goal" };
    if (kind === "agent-message") return { id: "agent" };
    if (kind === "team-message") return { id: "team" };
    if (kind === "webhook") return String(s.provider || "") === "github" ? { id: "github" } : { id: "webhook" };
    if (kind === "retry") return { id: "retry" };
    const form = String(s.form || "");
    if (form === "recall" || kind === "session-recall") return { id: "recall" };
    return { id: "inject", form };
  }
  var TRIGGER_TITLES = {
    job: () => t("\u540E\u53F0\u4EFB\u52A1\u72B6\u6001\u66F4\u65B0"),
    plugin: () => t("\u63D2\u4EF6\u72B6\u6001\u66F4\u65B0"),
    schedule: () => t("\u81EA\u52A8\u5316\u4EFB\u52A1"),
    goal: () => t("\u7EE7\u7EED\u6267\u884C\u76EE\u6807"),
    agent: () => t("\u6536\u5230\u4EFB\u52A1\u6D88\u606F"),
    team: () => t("\u6536\u5230\u56E2\u961F\u6D88\u606F"),
    github: () => t("\u6536\u5230 GitHub \u4E8B\u4EF6"),
    webhook: () => t("\u6536\u5230\u5916\u90E8\u4E8B\u4EF6"),
    retry: () => t("\u5DF2\u8BF7\u6C42\u7EE7\u7EED\u751F\u6210"),
    inject: () => t("\u4E0A\u4E0B\u6587\u6CE8\u5165"),
    recall: () => t("\u8DE8\u4F1A\u8BDD\u53EC\u56DE")
  };
  var SUBAGENT_TITLES = {
    completed: () => t("\u5B50\u667A\u80FD\u4F53\u5DF2\u5B8C\u6210\u4EFB\u52A1"),
    aborted: () => t("\u5B50\u667A\u80FD\u4F53\u4EFB\u52A1\u5DF2\u505C\u6B62"),
    error: () => t("\u5B50\u667A\u80FD\u4F53\u6267\u884C\u5931\u8D25"),
    "max-tokens": () => t("\u5B50\u667A\u80FD\u4F53\u5DF2\u8FBE token \u4E0A\u9650"),
    refusal: () => t("\u5B50\u667A\u80FD\u4F53\u62D2\u7EDD\u6267\u884C")
  };
  function noticeTitle(tg) {
    if (tg.id === "subagent") return (SUBAGENT_TITLES[tg.stopReason] || (() => t("\u5B50\u4EFB\u52A1\u72B6\u6001\u66F4\u65B0")))();
    return (TRIGGER_TITLES[tg.id] || TRIGGER_TITLES.inject)();
  }
  function noticeSummary(source) {
    const s = source && typeof source === "object" ? source : {};
    if (String(s.form || "") === "notice" && typeof s.summary === "string" && s.summary) {
      const t2 = String(s.summary);
      return t2.length > 120 ? t2.slice(0, 119) + "\u2026" : t2;
    }
    return "";
  }
  function producerLabel(source) {
    const s = source && typeof source === "object" ? source : {};
    if (typeof s.plugin === "string" && s.plugin) return s.plugin;
    const kind = String(s.kind || "");
    return kind && kind !== "user" ? kind : "";
  }
  function formatRunDuration(ms) {
    const total = Math.floor(Number(ms) || 0);
    if (total < 1e3) return "";
    const h = Math.floor(total / 36e5);
    const m = Math.floor(total % 36e5 / 6e4);
    const s = Math.floor(total % 6e4 / 1e3);
    const parts = [];
    if (h) parts.push(h + t("\u5C0F\u65F6"));
    if (m) parts.push(m + t("\u5206"));
    parts.push(s + t("\u79D2"));
    return parts.join("");
  }
  var pad2 = (n) => String(n).padStart(2, "0");
  function formatClock(time, now) {
    const d = new Date(Number(time) || 0);
    const n = new Date(Number(now) || Date.now());
    if (!Number(time) || isNaN(d.getTime())) return "";
    const clock = pad2(d.getHours()) + ":" + pad2(d.getMinutes());
    if (d.getFullYear() === n.getFullYear() && d.getMonth() === n.getMonth() && d.getDate() === n.getDate()) return clock;
    const md = d.getFullYear() === n.getFullYear() ? t("{m}\u6708{d}\u65E5").replace("{m}", String(d.getMonth() + 1)).replace("{d}", String(d.getDate())) : t("{y}\u5E74{m}\u6708{d}\u65E5").replace("{y}", String(d.getFullYear())).replace("{m}", String(d.getMonth() + 1)).replace("{d}", String(d.getDate()));
    return md + " " + clock;
  }
  function normalizeTurnReason(raw) {
    if (raw && typeof raw === "object") {
      if (raw.kind) return String(raw.kind);
      if (raw.reason && typeof raw.reason === "object" && raw.reason.kind) return String(raw.reason.kind);
      return "";
    }
    return String(raw ?? "");
  }
  function turnTailOf(reason, durationMs) {
    const r = normalizeTurnReason(reason);
    if (r === "interrupted" || r === "aborted" || r === "disposed" || r === "cancelled" || r === "canceled") {
      return { word: t("\u5DF2\u505C\u6B62"), duration: "", hint: "", cls: "stopped" };
    }
    if (r === "max-tokens") return { word: t("\u5DF2\u8FBE\u5230\u8F93\u51FA token \u4E0A\u9650"), duration: "", hint: t("\u56DE\u7B54\u88AB\u622A\u65AD\uFF0C\u5DF2\u6709\u8F93\u51FA\u4FDD\u7559\u5728\u5BF9\u8BDD\u4E2D\u3002\u53D1\u9001\u300C\u7EE7\u7EED\u300D\u53EF\u8BA9\u6A21\u578B\u63A5\u7740\u8F93\u51FA\u3002"), cls: "maxtok" };
    if (r && r !== "complete" && r !== "completed" && r !== "end_turn" && r !== "stop-sequence") {
      return { word: t("\u5904\u7406\u5931\u8D25"), duration: "", hint: "", cls: "failed" };
    }
    const dur = formatRunDuration(durationMs);
    return { word: dur ? t("\u5DF2\u5B8C\u6210\uFF0C\u7528\u65F6 ") + dur : t("\u5DF2\u5B8C\u6210"), duration: dur, hint: "", cls: "ok" };
  }
  function todoModel(todos) {
    const rows = (Array.isArray(todos) ? todos : []).filter((x) => x && typeof x.content === "string");
    const done = rows.filter((x) => x.status === "completed").length;
    const active = rows.find((x) => x.status === "in_progress");
    return { rows, done, total: rows.length, active: active ? active.content : "" };
  }
  var TOOL_TITLES = {
    bash: () => t("\u8FD0\u884C\u547D\u4EE4"),
    pwsh: () => t("\u8FD0\u884C\u547D\u4EE4"),
    read: () => t("\u8BFB\u53D6"),
    read_image: () => t("\u8BFB\u53D6\u56FE\u7247"),
    write: () => t("\u5199\u5165"),
    edit: () => t("\u7F16\u8F91"),
    apply_patch: () => t("\u7F16\u8F91"),
    grep: () => t("\u641C\u7D22\u6587\u4EF6\u5185\u5BB9"),
    glob: () => t("\u67E5\u627E\u6587\u4EF6"),
    web_search: () => t("\u7F51\u9875\u641C\u7D22"),
    web_fetch: () => t("\u7F51\u9875\u83B7\u53D6"),
    todo_write: () => t("\u66F4\u65B0\u4EFB\u52A1\u6E05\u5355"),
    ask_user_question: () => t("\u63D0\u95EE")
  };
  function toolTitle(name) {
    const fn = TOOL_TITLES[String(name || "").toLowerCase()];
    return fn ? fn() : "";
  }
  var SUMMARY_KEYS = ["file_path", "path", "command", "url", "query", "pattern", "note"];
  function toolSummary(name, argsRaw) {
    let obj = null;
    try {
      obj = JSON.parse(String(argsRaw || ""));
    } catch {
    }
    if (obj && typeof obj === "object" && !Array.isArray(obj)) {
      for (const k of SUMMARY_KEYS) {
        const v = obj[k];
        if (typeof v === "string" && v.trim()) return trunc(v.trim(), 90);
      }
    }
    const s = String(argsRaw || "").trim();
    return s ? trunc(s.split("\n")[0], 90) : "";
  }
  function toolStateOf(hasResult, errorObj, meta, name) {
    if (!hasResult) return "running";
    if (meta && typeof meta === "object" && meta.stopped === true) return "stopped";
    if (errorObj) return "error";
    const nm = String(name || "").toLowerCase();
    if ((nm === "bash" || nm === "pwsh") && meta && typeof meta === "object") {
      const code = Number(meta.exitCode);
      if (Number.isFinite(code) && code !== 0) return "error";
      if (typeof meta.signal === "string" && meta.signal) return "error";
    }
    return "ok";
  }
  function metaSummary(name, meta) {
    if (!meta || typeof meta !== "object") return "";
    const nm = String(name || "").toLowerCase();
    if (nm === "grep" || nm === "glob") {
      if (meta.shape === "paths" && Number.isFinite(Number(meta.total))) {
        return t("{shown} \u4E2A\u8DEF\u5F84").replace("{shown}", String(meta.truncated != null ? Number(meta.shown ?? meta.total) : Number(meta.total)));
      }
      if (Number.isFinite(Number(meta.total))) {
        if (Number(meta.total) === 0) return t("\u65E0\u7ED3\u679C");
        const files = Number.isFinite(Number(meta.files)) ? Number(meta.files) : null;
        return t("{shown} \u5904\u5339\u914D \xB7 {files} \u4E2A\u6587\u4EF6").replace("{shown}", String(Number(meta.shown ?? meta.total))).replace("{files}", files != null ? String(files) : "\u2014");
      }
      return "";
    }
    if (nm === "read" && Array.isArray(meta.lines) && meta.lines.length) {
      return t("{n} \u884C").replace("{n}", String(meta.lines.length));
    }
    if ((nm === "edit" || nm === "apply_patch" || nm === "write") && Array.isArray(meta.diffs)) {
      return meta.diffs.length ? t("{n} \u4E2A\u6587\u4EF6\u53D8\u66F4").replace("{n}", String(meta.diffs.length)) : "";
    }
    if (nm === "web_fetch" && Number.isFinite(Number(meta.statusCode))) {
      return "HTTP " + String(meta.statusCode);
    }
    return "";
  }
  function usageTotal(usage) {
    const u = usage && typeof usage === "object" ? usage : {};
    return ["inputTokens", "outputTokens", "cacheReadTokens", "cacheWriteTokens"].reduce((sum, k) => sum + (Number.isFinite(Number(u[k])) ? Number(u[k]) : 0), 0);
  }
  function usageDetail(usage) {
    const u = usage && typeof usage === "object" ? usage : {};
    const lines = [];
    if (Number.isFinite(Number(u.inputTokens))) lines.push(t("\u672A\u7F13\u5B58\u8F93\u5165") + " " + u.inputTokens);
    if (Number.isFinite(Number(u.cacheReadTokens))) lines.push(t("\u7F13\u5B58\u8BFB\u53D6") + " " + u.cacheReadTokens);
    if (Number.isFinite(Number(u.cacheWriteTokens))) lines.push(t("\u7F13\u5B58\u5199\u5165") + " " + u.cacheWriteTokens);
    if (Number.isFinite(Number(u.outputTokens))) lines.push(t("\u8F93\u51FA") + " " + u.outputTokens);
    if (Number.isFinite(Number(u.reasoningTokens))) lines.push(t("\uFF08\u5176\u4E2D\u63A8\u7406 {tokens}\uFF09").replace("{tokens}", String(u.reasoningTokens)));
    return lines.join(" \xB7 ");
  }
  function mapTranscript(events) {
    const nodes = [];
    const tools = /* @__PURE__ */ new Map();
    const order = [];
    const rows = Array.isArray(events) ? [...events] : [];
    rows.sort((a, b) => (Number(a && a.seq) || 0) - (Number(b && b.seq) || 0));
    const turnStarts = /* @__PURE__ */ new Map();
    let emittedSinceTurnStart = 0;
    let lastBoundaryTurnStart = false;
    for (const ev of rows) {
      if (!ev || typeof ev !== "object") continue;
      const type2 = String(ev.type || "");
      const data = ev.data && typeof ev.data === "object" ? ev.data : {};
      const seq2 = Number(ev.seq) || 0;
      const time = Number(ev.time) || 0;
      if (type2 === "turn/start") {
        turnStarts.set(Number(data.turn) || 0, time);
        emittedSinceTurnStart = 0;
        lastBoundaryTurnStart = true;
        continue;
      }
      if (type2 === "user/message") {
        const src = data.source && typeof data.source === "object" ? data.source : null;
        const kind = String(src && src.kind || "");
        if (kind === "compact-checkpoint") continue;
        const b = blocksOf(data.content);
        if (kind !== "user") {
          const tg = triggerOf(src);
          const waking = lastBoundaryTurnStart && emittedSinceTurnStart === 0;
          nodes.push({
            kind: "notice",
            seq: seq2,
            time,
            text: b.text,
            waking,
            trigger: tg,
            title: noticeTitle(tg),
            summary: noticeSummary(src),
            producer: producerLabel(src)
          });
          emittedSinceTurnStart += 1;
          lastBoundaryTurnStart = false;
          continue;
        }
        nodes.push({ kind: "user", seq: seq2, time, text: b.text, injected: false });
        emittedSinceTurnStart += 1;
        lastBoundaryTurnStart = false;
      } else if (type2 === "assistant/message") {
        const b = blocksOf(data.content);
        for (const tc of b.toolCalls) order.push(tc.id);
        if (!b.text && !b.reasoning && b.toolCalls.length) {
          emittedSinceTurnStart += 1;
          lastBoundaryTurnStart = false;
          continue;
        }
        nodes.push({ kind: "assistant", seq: seq2, time, text: b.text, reasoning: b.reasoning, usage: data.usage ?? null });
        emittedSinceTurnStart += 1;
        lastBoundaryTurnStart = false;
      } else if (type2 === "tool/call") {
        const id = String(data.callId || data.id || "");
        if (!tools.has(id)) tools.set(id, { seq: seq2, time, name: String(data.name || ""), args: typeof data.arguments === "string" ? data.arguments : JSON.stringify(data.arguments ?? data.args ?? ""), result: "", error: null, meta: null });
        order.push(id);
        emittedSinceTurnStart += 1;
        lastBoundaryTurnStart = false;
      } else if (type2 === "tool/result") {
        const id = String(data.callId || data.id || "");
        const t2 = tools.get(id);
        if (t2) {
          t2.result = resultTextOf(data);
          if (data.error && typeof data.error === "object") t2.error = { name: String(data.error.name || ""), code: String(data.error.code || "") };
          if (data.meta && typeof data.meta === "object") t2.meta = data.meta;
          if (!t2.error && data.message && typeof data.message === "object") {
            const bad = Array.isArray(data.message.content) && data.message.content.some((b) => b && typeof b === "object" && b.isError === true);
            if (bad) t2.error = { name: "", code: "" };
          }
        }
      } else if (type2 === "turn/end") {
        const turn = Number(data.turn) || 0;
        const startedAt = turnStarts.get(turn) || 0;
        nodes.push({ kind: "turntail", seq: seq2, time, reason: normalizeTurnReason(data.reason), durationMs: startedAt ? Math.max(0, time - startedAt) : 0 });
        emittedSinceTurnStart = 0;
        lastBoundaryTurnStart = false;
      } else if (type2 === "todo/write") {
        nodes.push({ kind: "todo", seq: seq2, time, todos: Array.isArray(data.todos) ? data.todos : [] });
        emittedSinceTurnStart += 1;
        lastBoundaryTurnStart = false;
      } else if (type2 === "llm/retry") {
        nodes.push({ kind: "retry", seq: seq2, time, data });
        emittedSinceTurnStart += 1;
        lastBoundaryTurnStart = false;
      } else if (type2 === "system/message" || type2 === "developer/message") {
        nodes.push({ kind: "system", seq: seq2, time, text: String(data.text ?? data.content ?? ""), dev: type2 === "developer/message" });
        emittedSinceTurnStart += 1;
        lastBoundaryTurnStart = false;
      }
    }
    const seenToolIds = /* @__PURE__ */ new Set();
    let lastTurnEndSeq = 0;
    for (const n of nodes) if (n.kind === "turntail" && Number(n.seq) > lastTurnEndSeq) lastTurnEndSeq = Number(n.seq);
    for (const id of order) {
      if (seenToolIds.has(id)) continue;
      seenToolIds.add(id);
      const t2 = tools.get(id);
      if (!t2) continue;
      const interrupted = !t2.result && lastTurnEndSeq > 0 && t2.seq < lastTurnEndSeq;
      const state = interrupted ? "stopped" : toolStateOf(!!t2.result, t2.error, t2.meta, t2.name);
      nodes.push({
        kind: "tool",
        seq: t2.seq,
        time: t2.time,
        name: t2.name || t("\u5DE5\u5177"),
        args: t2.args,
        result: t2.result,
        error: t2.error,
        meta: t2.meta,
        state,
        title: toolTitle(t2.name),
        summary: toolSummary(t2.name, t2.args),
        metaLine: metaSummary(t2.name, t2.meta)
      });
    }
    nodes.sort((a, b) => a.seq - b.seq);
    return nodes;
  }

  // src/screens/doc-chat.js
  var SESSIONS = /* @__PURE__ */ new Map();
  function faces2() {
    return typeof window !== "undefined" && window.__CC && window.__CC.chatFaces || null;
  }
  function extractYamlBlock(text2) {
    const s = String(text2 || "");
    const blocks = [...s.matchAll(/```(?:yaml|yml)?\s*\n([\s\S]*?)```/g)].map((m) => m[1]);
    for (let i = blocks.length - 1; i >= 0; i--) {
      const b = blocks[i].trim();
      if (/^(apiVersion|kind|metadata):/m.test(b)) return b;
    }
    return "";
  }
  function buildEditPrompt(ref, current, intent) {
    return t("\u8BF7\u4FEE\u6539\u5BF9\u8C61 ") + (ref ? ref.kind + "/" + (ref.scope || "global") + "/" + ref.name : "") + t("\uFF08\u5F53\u524D YAML \u58F0\u660E\u4E3A\u6743\u5A01\u6E90\uFF0C\u5168\u6587\u5982\u4E0B\uFF09\u3002\u4FEE\u6539\u610F\u56FE\uFF1A") + String(intent || "") + "\n\n```yaml\n" + String(current || "") + "\n```\n" + t("\u8F93\u51FA\u8981\u6C42\uFF1A\u7ED9\u51FA\u4FEE\u6539\u540E\u7684\u5B8C\u6574 YAML \u58F0\u660E\uFF0C\u653E\u5728\u4E00\u4E2A ```yaml \u4EE3\u7801\u5757\u91CC\uFF1B\u4E0D\u6539\u65E0\u5173\u5B57\u6BB5\uFF1B\u89E3\u91CA\u653E\u4EE3\u7801\u5757\u5916\u3002");
  }
  function transcriptRows(events) {
    const rows = [];
    (events || []).forEach((e) => {
      if (!e || !e.type) return;
      const d = e.data || {};
      if (e.type === "user/message" && d.source && d.source.kind === "user") {
        const text2 = (d.content || []).filter((b) => b && b.type === "text").map((b) => b.text).join("\n");
        rows.push({ kind: "user", seq: e.seq, text: text2 });
      } else if (e.type === "assistant/message") {
        const msg = d.message || d;
        const blocks = Array.isArray(msg.content) ? msg.content : [];
        const reasoning = blocks.filter((b) => b.type === "reasoning").map((b) => b.text).join("\n");
        const text2 = blocks.filter((b) => b.type === "text").map((b) => b.text).join("\n");
        rows.push({ kind: "assistant", seq: e.seq, text: text2, reasoning });
      } else if (e.type === "turn/end") {
        rows.push({ kind: "tail", seq: e.seq, tail: turnTailOf(d.reason, 0) });
      }
    });
    return rows;
  }
  function rowsHtmlOf(rows) {
    const list2 = rows || [];
    if (!list2.length) return '<div class="small faint">' + esc(t("\u5C1A\u65E0\u5BF9\u8BDD\u2014\u2014\u4E0B\u65B9\u8F93\u5165\u4FEE\u6539\u610F\u56FE\u53D1\u9001\u3002")) + "</div>";
    return list2.map((r) => {
      if (r.kind === "user") {
        return '<div class="dc-row dc-user"><div class="dc-bubble">' + esc(r.text) + "</div></div>";
      }
      if (r.kind === "assistant") {
        const rfold = r.reasoning ? '<details class="chat-fold dc-fold"><summary><span class="ex-chev">\u25B8</span>' + esc(t("\u601D\u8003")) + '</summary><div class="ex-body">' + esc(r.reasoning) + "</div></details>" : "";
        return '<div class="dc-row dc-assistant">' + rfold + (r.text ? '<div class="dc-text">' + esc(r.text) + "</div>" : "") + "</div>";
      }
      if (r.kind === "tail" && r.tail) {
        return '<div class="dc-row dc-tail"><span class="dc-tail-' + esc(r.tail.cls || "ok") + '">' + esc(r.tail.word || "") + "</span></div>";
      }
      return "";
    }).join("");
  }
  function panelHtml(state) {
    const s = state || {};
    return '<div class="dc-meta">' + (s.refLine || "") + '</div><div class="dc-transcript" data-dc-transcript>' + rowsHtmlOf(s.rows) + '</div><div data-dc-diff></div><div class="dc-composer"><textarea class="inp" data-dc-text rows="2" placeholder="' + esc(t("\u4FEE\u6539\u610F\u56FE\uFF08\u5982\uFF1A\u628A standard \u6A21\u578B\u6362\u6210 reasoning-tier\uFF09\u2014\u2014\u53D1\u9001\u540E\u52A9\u624B\u56DE\u4FEE\u6539\u540E YAML\uFF0C\u5BF9\u6BD4\u786E\u8BA4\u5373 Apply")) + '"></textarea><div class="dc-acts"><button type="button" class="btn btn-chip sm" data-dc-preview>' + esc(t("\u5BF9\u6BD4\u5E76 Apply")) + '</button><button type="button" class="btn btn-chip sm" data-dc-send>' + esc(t("\u53D1\u9001")) + '</button><button type="button" class="btn btn-ghost sm" data-dc-stop hidden>' + esc(t("\u505C\u6B62")) + '</button><span class="small faint" data-dc-msg>' + esc(s.note || "") + "</span></div></div>";
  }
  function openDocChat(opts) {
    const o = opts || {};
    const f = faces2();
    const node = document.createElement("div");
    node.className = "doc-chat";
    if (!f || typeof f.send !== "function") {
      node.innerHTML = '<div class="ob-empty">' + esc(t("\u5BF9\u8BDD\u901A\u8DEF\u672A\u6CE8\u5165\uFF08chatfaces.js \u7F3A\u5E2D\uFF09\u2014\u2014\u539F\u5730\u5BF9\u8BDD\u4FEE\u6539\u4E0D\u53EF\u7528\uFF1B\u53EF\u5728\u300C\u539F\u6587\u300D\u89C6\u56FE\u624B\u7F16\uFF0C\u6216\u7ECF\u4F1A\u8BDD\u5C4F\u5BF9\u8BDD\u3002")) + "</div>";
      htmlModal({ title: o.title || t("\u5BF9\u8BDD\u4FEE\u6539"), wide: true, node, holder: document.body });
      return;
    }
    const painted = /* @__PURE__ */ new Set();
    let sessionId = SESSIONS.get(o.refKey || "") || "";
    let running = false;
    let current = String(o.current || "");
    const state = { rows: [], refLine: "", note: "" };
    const park = document.createElement("div");
    node.innerHTML = panelHtml(state);
    htmlModal({ title: o.title || t("\u5BF9\u8BDD\u4FEE\u6539"), wide: true, node, holder: park }).then(() => {
      clearInterval(timer);
    });
    const paintRows = () => {
      const box = node.querySelector("[data-dc-transcript]");
      if (box) box.innerHTML = rowsHtmlOf(state.rows);
      box && (box.scrollTop = box.scrollHeight);
    };
    const refresh3 = async () => {
      if (!sessionId) return;
      const r = await f.transcript(sessionId, 40);
      if (r.status !== 200 || !Array.isArray(r.data && r.data.events)) return;
      state.rows = transcriptRows(r.data.events);
      paintRows();
      const st = await f.state(sessionId);
      running = !!(st.status === 200 && st.data && st.data.running === true);
      const send2 = node.querySelector("[data-dc-send]");
      const stop = node.querySelector("[data-dc-stop]");
      if (send2) send2.disabled = running;
      if (stop) stop.hidden = !running;
    };
    const ensureSession = async () => {
      if (sessionId) return sessionId;
      const r = await f.create();
      if (r.status === 200 && r.data && r.data.sessionId) {
        sessionId = String(r.data.sessionId);
        SESSIONS.set(o.refKey || "", sessionId);
        return sessionId;
      }
      throw new Error(t("\u65B0\u4F1A\u8BDD\u5931\u8D25\uFF08") + String(r.status) + t("\uFF09"));
    };
    const send = async () => {
      const ta = node.querySelector("[data-dc-text]");
      const intent = String(ta && ta.value || "").trim();
      const msg = node.querySelector("[data-dc-msg]");
      if (!intent) {
        if (msg) msg.textContent = t("\u8BF7\u5148\u5199\u4FEE\u6539\u610F\u56FE");
        return;
      }
      if (running) {
        if (msg) msg.textContent = t("\u4E0A\u4E00\u8F6E\u4ECD\u5728\u8FD0\u884C\u2014\u2014\u53EF\u70B9\u505C\u6B62");
        return;
      }
      try {
        const sid = await ensureSession();
        const first = !state.rows.length;
        const text2 = first ? buildEditPrompt(o.ref, current, intent) : intent;
        const r = await f.send(sid, text2, "queue");
        if (r.status !== 200) {
          if (msg) msg.textContent = t("\u53D1\u9001\u5931\u8D25\uFF1A") + String(r.data && r.data.error || r.status);
          return;
        }
        if (ta) ta.value = "";
        await refresh3();
      } catch (e) {
        if (msg) msg.textContent = String(e && e.message || e);
      }
    };
    const showDiff = () => {
      const lastAssistant = [...state.rows].reverse().find((r) => r.kind === "assistant" && r.text);
      const modified = (o.extractApply || extractYamlBlock)(lastAssistant ? lastAssistant.text : "");
      const diffBox = node.querySelector("[data-dc-diff]");
      if (!modified) {
        diffBox.innerHTML = '<div class="small faint" style="margin-top:8px">' + esc(t("\u6700\u8FD1\u56DE\u590D\u672A\u542B ```yaml \u4EE3\u7801\u5757\u2014\u2014\u8BF7\u8BA9\u52A9\u624B\u7ED9\u51FA\u4FEE\u6539\u540E\u7684\u5B8C\u6574 YAML \u58F0\u660E\u3002")) + "</div>";
        return;
      }
      const hunks = hunksOf(diffLines2(current, modified));
      diffBox.innerHTML = '<div class="card" style="margin-top:10px"><div class="card-h"><span class="h-title">' + esc(t("\u4FEE\u6539\u9884\u89C8\uFF08") + hunks.length + t(" \u5904\u53D8\u66F4\uFF09")) + '</span><span class="h-right"><button type="button" class="btn btn-primary sm" data-dc-apply>' + esc(t("\u6279\u51C6\uFF08Apply\uFF09")) + '</button> <button type="button" class="btn btn-secondary sm" data-dc-dismiss>' + esc(t("\u653E\u5F03")) + '</button></span></div><div class="card-b" style="max-height:300px;overflow:auto">' + renderDiff(diffLines2(current, modified)) + "</div></div>";
      diffBox.querySelector("[data-dc-dismiss]").addEventListener("click", () => {
        diffBox.innerHTML = "";
      });
      diffBox.querySelector("[data-dc-apply]").addEventListener("click", async () => {
        const msg = node.querySelector("[data-dc-msg]");
        if (msg) msg.textContent = t("\u6B63\u5728 Apply\u2026");
        try {
          const face = await o.ctx.api.object(o.ref.kind, o.ref.scope, o.ref.name);
          const rv = face.ok && face.data && face.data.resourceVersion != null ? Number(face.data.resourceVersion) : o.ref.rv || 0;
          const r = await o.ctx.api.apply({
            kind: o.ref.kind,
            scope: o.ref.scope || "global",
            name: o.ref.name,
            resourceVersion: rv,
            content: modified,
            fieldManager: "console:doc-chat",
            reasonRef: "doc-chat:" + o.ref.name
          });
          if (r.ok) {
            current = modified;
            diffBox.innerHTML = "";
            if (msg) msg.textContent = t("\u5DF2 Apply\uFF08rv=") + ((r.data && r.data.rv) != null ? r.data.rv : "?") + t("\uFF09");
            toast(t("\u5BF9\u8BDD\u4FEE\u6539\u5DF2 Apply"), "accent");
            if (typeof o.onApplied === "function") o.onApplied(modified);
          } else {
            if (msg) msg.textContent = t("Apply \u5931\u8D25\uFF1A") + String(r.error || r.status) + (String(r.status) === "409" ? t("\uFF08\u7248\u672C\u5DF2\u53D8\u2014\u2014\u91CD\u8BD5\u5C06\u53D6\u65B0\u7248\u672C\uFF09") : "");
            toast(t("Apply \u5931\u8D25\uFF08") + String(r.error || r.status) + "\uFF09", "error");
          }
        } catch (e) {
          if (msg) msg.textContent = String(e && e.message || e);
        }
      });
    };
    node.querySelector("[data-dc-send]").addEventListener("click", send);
    node.querySelector("[data-dc-stop]").addEventListener("click", async () => {
      if (sessionId) await f.stop(sessionId);
    });
    node.querySelector("[data-dc-text]").addEventListener("keydown", (ev) => {
      if (ev.key === "Enter" && !ev.shiftKey) {
        ev.preventDefault();
        send();
      }
    });
    node.querySelector("[data-dc-preview]").addEventListener("click", showDiff);
    const timer = setInterval(() => {
      refresh3().catch(() => {
      });
    }, 1800);
    refresh3().catch(() => {
    });
  }

  // src/screens/workflows.js
  function wfAnchorOfRoute(route) {
    const r = String(route || "");
    if (!r) return "";
    if (r.startsWith("workflows/")) return r.slice("workflows/".length);
    if (/-workflow$/.test(r)) return "core-workflow";
    return "";
  }
  var RUN_CLS = { running: "acc", done: "ok", ok: "ok", success: "ok", converged: "ok", failed: "esc", error: "esc", canceled: "mute", queued: "mute", pending: "mute" };
  function isTerminalRun(st) {
    return ["done", "converged", "failed", "canceled", "ok", "success"].includes(String(st || ""));
  }
  var REG_CLS = { running: "acc", active: "acc", suspended: "warn", awaitingapproval: "warn", failed: "esc", error: "esc", done: "ok", ok: "ok", idle: "mute", registered: "mute" };
  var PH_CLS = { awaitingapproval: "warn", failed: "esc", done: "ok", executing: "acc", running: "acc" };
  function runChip(r) {
    const st = String(r && (r.status || r.state || r.phase) || t("\u672A\u77E5"));
    return chip(st, RUN_CLS[st] || "mute");
  }
  function phChip(ph) {
    const p = String(ph || "");
    return chip(p || "\u2014", PH_CLS[p.toLowerCase()] || "mute");
  }
  function wfDesignRoute() {
    return "designdoc/" + encodeURIComponent("core_design") + "/" + encodeURIComponent("concepts/workflow.md");
  }
  function realTplCardHtml(o) {
    const scope = o.scope || "global";
    return '<div class="card"><div class="card-h"><span class="h-title mono">' + esc(o.name) + '</span><span class="h-right">' + chip("Template", "acc") + " " + chip("r" + String(o.rv || 0), "mute") + '</span></div><div class="card-b"><div class="small mono faint">scope: ' + esc(scope) + '</div><div class="dh-actions" style="margin-top:8px">' + chipGo(t("\u5BF9\u8C61\u5DE5\u4F5C\u53F0"), objRoute("Template", scope, o.name), "acc", true, t("\u67E5\u770B Template \u58F0\u660E")) + " " + chipGo(t("\u7F16\u8F91\u5668"), docRoute("Template", scope, o.name), "mute", true, t("\u67E5\u770B\u6B63\u6587")) + "</div></div></div>";
  }
  function templatesSectionHtml(ctx2) {
    const tpls = ctx2.store.refs("Template");
    const bugWf = ctx2.store.refs("Workflow").find((w) => w && w.name === "bug-lifecycle");
    const bugRuns = ctx2.store.runs().filter((r) => r && String(r.workflow || r.name || "") === "bug-lifecycle");
    const bugBlock = bugWf || bugRuns.length ? '<div class="card" style="margin-bottom:10px"><div class="card-h"><span class="h-title mono">bug-lifecycle</span><span class="h-right">' + chip("Workflow", "acc") + " " + chip(bugRuns.length + t(" \u6B21\u8FD0\u884C"), bugRuns.length ? "acc" : "mute") + '</span></div><div class="card-b"><div class="dh-actions">' + chipGo(t("\u5DE5\u4F5C\u6D41\u8BE6\u60C5"), "wf/bug-lifecycle", "acc", true) + " " + chipGo(t("\u6587\u6863\u5E93 BugDoc"), "docs", "mute", true) + "</div></div></div>" : "";
    const tplBlock = tpls.length ? '<div class="grid cols-2">' + tpls.map(realTplCardHtml).join("") + "</div>" : '<div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u6682\u65E0\u6D41\u7A0B\u6A21\u677F\uFF08Template \u5BF9\u8C61\u5728\u518C\u540E\u6B64\u5904\u663E\u793A\uFF09\u3002")) + " " + chipGo(t("\u8BBE\u8BA1\u89C1 workflow.md \xA79b"), wfDesignRoute(), "mute", true, t("\u8BBE\u8BA1\u6587\u6863\u5DE5\u4F5C\u53F0")) + "</div></div></div>";
    return bugBlock + tplBlock;
  }
  function listScreen(el, ctx2) {
    const state = { rows: [], filter: "", err: "" };
    const snapRuns = ctx2.store.runs().slice(0, 8);
    const runRowHtml = (r) => {
      const wf = String(r && (r.workflow || r.name) || "");
      const no = r && (r.run != null ? r.run : r.no != null ? r.no : "");
      if (!wf && (no == null || no === "")) return "";
      const started = r && (r.startedAt || r.at || r.ts);
      const invChips = String(r && r.involved || "").split(",").map((x) => x.trim()).filter(Boolean).slice(0, 3).map((x) => {
        const rt = involvedRoute(x);
        return rt ? chipGo(trunc(x, 36), rt, "mute", true, t("\u53BB\u5BF9\u8C61\u5DE5\u4F5C\u53F0")) : '<span class="chip mute tiny">' + esc(trunc(x, 36)) + "</span>";
      }).join(" ");
      const go = wf && no != null && no !== "" ? "wf/" + encodeURIComponent(wf) + "/r/" + encodeURIComponent(String(no)) : "";
      return "<tr" + (go ? ' data-go="' + esc(go) + '" style="cursor:pointer"' : "") + '><td class="mono small">#' + esc(String(no != null && no !== "" ? no : "\u2014")) + '</td><td class="mono small">' + esc(wf || "\u2014") + "</td><td>" + phChip(r && r.phase) + "</td><td>" + runChip(r) + '</td><td class="small mono">' + esc(relAge(started) || (started ? String(started).slice(0, 16) : "\u2014")) + '</td><td class="small">' + (invChips || '<span class="faint">\u2014</span>') + "</td></tr>";
    };
    const paintRuns = () => {
      const slot = el.querySelector("[data-wf-runs]");
      if (!slot) return;
      const seen = /* @__PURE__ */ new Set();
      const merged = [];
      for (const r of state.rows.concat(snapRuns)) {
        const k = String(r && (r.workflow || r.name) || "") + "#" + String(r && (r.run != null ? r.run : r.no != null ? r.no : ""));
        if (seen.has(k)) continue;
        seen.add(k);
        merged.push(r);
      }
      const shown = (state.filter ? merged.filter((r) => String(r && (r.workflow || r.name) || "") === state.filter) : merged).slice(0, 50);
      slot.innerHTML = '<div class="card"><div class="card-b" style="overflow-x:auto"><div style="margin-bottom:6px">' + chip(merged.length + t(" \u6761\u8FD0\u884C\u8BB0\u5F55"), "mute") + " " + (state.filter ? chipGo(t("\u6E05\u9664\u8FC7\u6EE4\uFF1A") + state.filter, "workflows", "warn", true, t("\u6E05\u9664\u8FC7\u6EE4\u5E76\u56DE\u5168\u8868")) : "") + "</div>" + (state.err ? '<div class="small faint" style="margin-bottom:6px">' + esc(t("\u8FD0\u884C\u8BB0\u5F55\u6682\u4E0D\u53EF\u8BFB\uFF08")) + esc(state.err) + esc(t("\uFF09\u2014\u2014\u4E0B\u8868\u4EC5\u6982\u89C8\u884C\u3002")) + "</div>" : "") + (shown.length ? '<table class="tbl"><thead><tr><th>#</th><th>' + esc(t("\u5DE5\u4F5C\u6D41")) + "</th><th>" + esc(t("\u76F8\u4F4D")) + "</th><th>" + esc(t("\u72B6\u6001")) + "</th><th>" + esc(t("\u5F00\u59CB")) + "</th><th>" + esc(t("\u6D89\u53CA")) + "</th></tr></thead><tbody>" + shown.map(runRowHtml).join("") + '</tbody></table><div class="small faint" style="margin-top:6px">' + esc(t("\u884C\u70B9\u51FB\u8FDB\u8FD0\u884C\u8BE6\u60C5\uFF1Binvolved chip=\u5BF9\u8C61\u5DE5\u4F5C\u53F0\u6DF1\u94FE\u3002")) + "</div>" : '<div class="small faint">' + esc(state.filter ? t("\u8BE5 workflow \u7A97\u53E3\u5185\u6682\u65E0\u8FD0\u884C\u884C\u3002") : t("\u6682\u65E0\u8FD0\u884C\u8BB0\u5F55\u3002"))) + "</div></div>";
    };
    el.innerHTML = '<div class="view-inner">' + screenHead(
      t("\u5DE5\u4F5C\u6D41"),
      t("\u4E09\u533A=\u5728\u518C\u6CE8\u518C\u8868\uFF08GET /v1/workflows \u73B0\u8BFB\uFF09\u222A \u8FD0\u884C\u8BB0\u5F55\uFF08GET /v1/workflows/runs+\u5FEB\u7167\u805A\u5408\u5E76\u663E\uFF09\u222A \u6A21\u677F\uFF08\xA79b \u8BBE\u8BA1\u5361\uFF09\u3002"),
      chipGo(t("\u8BBE\u8BA1\u6587\u6863"), wfDesignRoute(), "mute", true, t("workflow.md \xA79b"))
    ) + '<h2 class="sec-title">' + esc(t("\u5728\u518C\u5DE5\u4F5C\u6D41")) + '</h2><div data-wf-reg><div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u6CE8\u518C\u8868\u73B0\u8BFB\u4E2D\u2026")) + '</div></div></div></div><h2 class="sec-title" style="margin-top:18px">' + esc(t("\u8FD0\u884C\u8BB0\u5F55")) + '</h2><div data-wf-runs><div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u8FD0\u884C\u8BB0\u5F55\u73B0\u8BFB\u4E2D\u2026")) + '</div></div></div></div><h2 class="sec-title" style="margin-top:18px">' + esc(t("\u6A21\u677F")) + "</h2>" + templatesSectionHtml(ctx2) + "</div>";
    paintRuns();
    (async () => {
      const stamp = typeof location !== "undefined" ? location.hash : "";
      const slot = el.querySelector("[data-wf-reg]");
      const r = await ctx2.api.workflows();
      if (!slot || !slot.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
      if (!r.ok) {
        slot.innerHTML = '<div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u5DE5\u4F5C\u6D41\u6E05\u5355\u6682\u4E0D\u53EF\u8BFB\uFF08")) + esc(String(r.error || r.status)) + esc(t("\uFF09\uFF0C\u8BF7\u5237\u65B0\u91CD\u8BD5\u3002")) + "</div></div></div>";
        return;
      }
      const data = r.data && typeof r.data === "object" ? r.data : {};
      const rows = Array.isArray(data.workflows) ? data.workflows : [];
      if (!rows.length) {
        slot.innerHTML = '<div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u6682\u65E0\u5728\u518C\u5DE5\u4F5C\u6D41\u3002")) + "</div></div></div>";
        return;
      }
      slot.innerHTML = '<div class="card"><div class="card-b" style="overflow-x:auto">' + chip((data.count != null ? data.count : rows.length) + t(" \u6761\u5728\u518C"), "acc") + " " + (data.scanCapped ? chip(t("\u626B\u63CF\u622A\u65AD\uFF08\u7A97\u53E3\u5185 ") + String(data.scanned != null ? data.scanned : "?") + t(" \u6761\uFF09"), "warn") : "") + '<table class="tbl" style="margin-top:8px"><thead><tr><th>' + esc(t("\u540D\u79F0")) + "</th><th>" + esc(t("\u72B6\u6001")) + "</th><th>" + esc(t("\u8FDE\u63A5")) + "</th><th>" + esc(t("\u6700\u8FD1")) + "</th><th></th></tr></thead><tbody>" + rows.map((w) => {
        const name = String(w && w.name || "");
        const ex = w && w.execution || {};
        const exState = String(ex.state || "");
        const connTxt = String(ex.connector || "") + (ex.connectorScope ? "/" + String(ex.connectorScope) : "");
        const connCell = connTxt ? chipGo(connTxt, "connectors", "mute", true, t("\u8FDE\u63A5\u5C42\u2014\u2014\u67E5\u770B\u7ED1\u5B9A connector")) : '<span class="faint">\u2014</span>';
        const latest = w && w.latest;
        const latestTxt = typeof latest === "string" ? latest : String(latest && (latest.summary || latest.note || latest.status || latest.state) || "");
        return '<tr><td class="mono small">' + (name ? '<span class="lk" data-go="wf/' + encodeURIComponent(name) + '">' + esc(name) + "</span>" : "\u2014") + "</td><td>" + chip(exState || "\u2014", REG_CLS[exState.toLowerCase()] || "mute") + '</td><td class="mono small">' + connCell + '</td><td class="small">' + esc(trunc(latestTxt || "\u2014", 60)) + '</td><td style="text-align:right">' + (name ? '<span class="chip mute tiny click" data-wf-runfilter="' + esc(name) + '" title="' + esc(t("\u8FC7\u6EE4\u672C\u9875\u8FD0\u884C\u8BB0\u5F55")) + '">' + esc(t("\u8FD0\u884C\u8BB0\u5F55")) + "</span>" : "") + "</td></tr>";
      }).join("") + "</tbody></table></div></div>";
      slot.querySelectorAll("[data-wf-runfilter]").forEach((c) => {
        c.addEventListener("click", () => {
          state.filter = String(c.getAttribute("data-wf-runfilter") || "");
          paintRuns();
          const rs = el.querySelector("[data-wf-runs]");
          if (rs && rs.scrollIntoView) rs.scrollIntoView({ behavior: "smooth", block: "start" });
        });
      });
    })();
    (async () => {
      const stamp = typeof location !== "undefined" ? location.hash : "";
      const r = await ctx2.api.workflowRuns();
      if (!el.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
      if (!r.ok) {
        state.err = String(r.error || r.status);
      } else {
        const data = r.data && typeof r.data === "object" ? r.data : {};
        state.rows = Array.isArray(data.runs) ? data.runs : [];
      }
      paintRuns();
    })();
  }
  var WF_ALIASES = {
    "bug-workflow": "core-system-developer-loop",
    "bug-lifecycle": "core-system-developer-loop",
    "core-workflow": "core-system-developer-loop",
    "dev-delivery": "core-system-developer-loop",
    "apply-workflow": "default-task",
    "simple-shot": "default-task",
    "watch-only": "default-task"
  };
  function wfApiRows(res) {
    return res && res.ok && res.data && Array.isArray(res.data.workflows) ? res.data.workflows : [];
  }
  function wfApiRuns(res) {
    return res && res.ok && res.data && Array.isArray(res.data.runs) ? res.data.runs : [];
  }
  function wfApiTasks(res) {
    return res && res.ok && res.data && Array.isArray(res.data.tasks) ? res.data.tasks : [];
  }
  function wfTaskIndex(tasks) {
    const m = /* @__PURE__ */ new Map();
    for (const t2 of tasks) if (t2 && t2.name) m.set(String(t2.scope || "global") + "/" + String(t2.name), t2);
    return m;
  }
  function wfTaskKey(involved) {
    const m = /^Task\/([^/]+)\/([^/]+)/.exec(String(involved || ""));
    return m ? m[1] + "/" + m[2] : "";
  }
  var SESSION_MATCH_WINDOW_MS = 10 * 60 * 1e3;
  function wfSessionAnchor(run, task, sessionRows, hist = null) {
    const hs = String(hist && hist.sessionId || "");
    if (/^session[-_]/i.test(hs)) return { sid: hs, label: t("\u8282\u70B9\u4F1A\u8BDD") };
    const direct = String(run && run.sessionId || "");
    if (/^session[-_]/i.test(direct)) return { sid: direct, label: t("\u8FD0\u884C\u4F1A\u8BDD") };
    const key = wfTaskKey(run && run.involved);
    const candidates = (Array.isArray(sessionRows) ? sessionRows : []).filter((x) => x && x.taskId === key && /^session[-_]/i.test(String(x.sessionId || "")));
    if (!candidates.length) return { sid: "", label: t("\u4EFB\u52A1\u4F1A\u8BDD") };
    const started = Date.parse(String(run && (run.startedAt || run.at || run.ts) || ""));
    if (!Number.isFinite(started)) {
      return candidates.length === 1 ? { sid: String(candidates[0].sessionId), label: t("\u4EFB\u52A1\u4F1A\u8BDD") } : { sid: "", label: t("\u4EFB\u52A1\u4F1A\u8BDD") };
    }
    const ranked = candidates.map((x) => ({
      x,
      d: Number.isFinite(Number(x.startedAt)) ? Math.abs(Number(x.startedAt) - started) : Number.POSITIVE_INFINITY
    })).sort((a, b) => a.d - b.d);
    if (!ranked.length || ranked[0].d > SESSION_MATCH_WINDOW_MS) return { sid: "", label: t("\u4EFB\u52A1\u4F1A\u8BDD") };
    return { sid: String(ranked[0].x.sessionId), label: t("\u4EFB\u52A1\u4F1A\u8BDD") };
  }
  function wfSessionChip(sid, label) {
    const s = String(sid || "");
    if (!s) return '<span class="small faint">' + esc(t("\u4F1A\u8BDD\u672A\u8BB0\u5F55")) + "</span>";
    const isSession = /^session[-_]/i.test(s);
    if (!isSession) return '<span class="small faint">' + esc((label ? t(label) + " " : "") + t("\u4EFB\u52A1\u951A\uFF08\u975E sessionId\uFF09") + " " + trunc(s, 24)) + "</span>";
    return chipGo((label ? t(label) + " " : "") + trunc(s, 24), "session/" + encodeURIComponent(s), "acc", true, s);
  }
  function wfNodeStats(runs, phases) {
    const stats = /* @__PURE__ */ new Map();
    for (const p of phases) stats.set(p, { cur: 0, visited: 0, failed: 0 });
    for (const r of runs) {
      const cur = String(r.phase || "");
      if (stats.has(cur)) stats.get(cur).cur++;
      const seen = new Set((Array.isArray(r.history) ? r.history : []).map((h) => String(h && h.to || "")));
      for (const p of seen) if (stats.has(p)) stats.get(p).visited++;
      if (String(r.status || "") === "failed" && stats.has("Failed")) stats.get("Failed").failed++;
    }
    return stats;
  }
  function wfNodeGraphHtml(wf, runs) {
    const phases = (Array.isArray(wf && wf.phases) ? wf.phases : []).map(String);
    const failed = runs.filter((r) => String(r.status || "") === "failed").length;
    const list2 = phases.concat(failed && !phases.includes("Failed") ? ["Failed"] : []);
    if (!list2.length) return '<div class="small faint">' + esc(t("\u8BE5 workflow \u672A\u6295\u5F71\u76F8\u4F4D\u94FE\u3002")) + "</div>";
    const raw = String(wf && wf.rawYaml || "");
    if (raw) {
      try {
        const doc = loadYaml(raw) || {};
        const phasesRaw = Array.isArray(doc && doc.spec && doc.spec.phases) ? doc.spec.phases : [];
        const transitions = Array.isArray(doc && doc.spec && doc.spec.transitions) ? doc.spec.transitions : [];
        if (transitions.length) {
          const phasesRich = phasesRaw.map((n) => typeof n === "string" ? { name: n, phase: n } : n).filter((n) => n && (n.phase || n.name));
          const graph = wfGraphHtml(phasesRich, transitions);
          if (graph) {
            return '<div class="small faint" style="margin-bottom:6px">' + esc(t("\u76F8\u4F4D\u56FE\u2014\u2014\u4EA4\u53C9\u51FA\u53E3=\u5224\u65AD\u5206\u53C9\u8FB9\uFF08\u8FB9\u4E0A\u8BCD\u5F62=\u5224\u65AD\u8BCD\uFF1Bjudge \u4F1A\u8BDD [decision:\u8BCD] \u6807\u8BB0\u6216\u5224\u5B9A\u57DF connectops-jev \u56DE\u4F20\u70B9\u706B\uFF0C\u672A\u547D\u4E2D=advance \u515C\u5E95\uFF09\u3002")) + '</div><div style="overflow-x:auto;padding-bottom:4px">' + graph + '</div><div class="small faint">' + esc(t("\u70B9\u51FB\u8282\u70B9\uFF1A\u5230\u8FBE\u8BE5\u8282\u70B9\u7684\u8FD0\u884C/\u4EFB\u52A1\u4E0E\u771F\u5B9E\u4F1A\u8BDD\u951A\u3002")) + '</div><div data-wf-node-drawer class="wf-node-drawer"></div>';
          }
        }
      } catch (e) {
      }
    }
    const stats = wfNodeStats(runs, list2);
    const body = list2.map((p) => {
      const s = stats.get(p) || { cur: 0, visited: 0, failed: 0 };
      const cls = "pg-node" + (s.cur ? " active" : "") + (s.visited || s.failed ? " done" : "") + (p === "Failed" ? " fail" : "");
      return '<div class="' + cls + '" data-wf-node="' + esc(p) + '" role="button"><div class="pg-dot"></div><div class="pg-name">' + esc(p) + "</div>" + (s.cur ? '<span class="chip acc tiny">' + esc(t("\u5F53\u524D") + " " + s.cur) + "</span>" : "") + (s.visited ? '<span class="chip mute tiny">' + esc(t("\u5230\u8FC7") + " " + s.visited) + "</span>" : "") + "</div>";
    }).join('<span class="pg-edge">\u2192</span>');
    return '<div class="pg-wrap">' + body + '</div><div class="small faint">' + esc(t("\u70B9\u51FB\u8282\u70B9\uFF1A\u5230\u8FBE\u8BE5\u8282\u70B9\u7684\u8FD0\u884C/\u4EFB\u52A1\u4E0E\u771F\u5B9E\u4F1A\u8BDD\u951A\u3002")) + '</div><div data-wf-node-drawer class="wf-node-drawer"></div>';
  }
  function bindWfNodeGraph(el, runs, taskIdx, wfName, sessionRows = []) {
    el.querySelectorAll("[data-wf-node],[data-graph-node]").forEach((n) => {
      n.addEventListener("click", () => {
        const ph = String(n.getAttribute("data-wf-node") || n.getAttribute("data-graph-node") || "");
        const drawer = el.querySelector("[data-wf-node-drawer]");
        if (!drawer) return;
        const hits = runs.filter((r) => String(r.phase || "") === ph || (Array.isArray(r.history) ? r.history : []).some((h) => String(h && h.to || "") === ph || String(h && h.from || "") === ph));
        const rows = hits.length ? hits.map((r) => {
          const no = r.run != null ? r.run : r.no != null ? r.no : "";
          const task = taskIdx.get(wfTaskKey(r.involved));
          const hist = (Array.isArray(r.history) ? r.history : []).find((h) => String(h && h.to || "") === ph || String(h && h.from || "") === ph);
          const anchor = wfSessionAnchor(r, task, sessionRows, hist);
          return '<div class="mini-row"><div class="mr-main"><div><b class="mono">run #' + esc(String(no)) + "</b> " + runChip(r) + '</div><div class="mr-sub small">' + esc(task ? String(task.title || task.name || "") : String(r.involved || "\u2014")) + '</div></div><div class="mr-side">' + wfSessionChip(anchor.sid, anchor.label) + " " + chipGo(t("\u8FD0\u884C"), "wf/" + encodeURIComponent(wfName) + "/r/" + encodeURIComponent(String(no)), "mute", true) + "</div></div>";
        }).join("") : '<div class="small faint">' + esc(t("\u6682\u65E0\u8FD0\u884C\u5230\u8FBE\u8BE5\u8282\u70B9\u3002")) + "</div>";
        drawer.innerHTML = '<div class="card"><div class="card-h"><span class="h-title">' + esc(ph) + '</span><span class="h-right">' + esc(t("\u5230\u8FBE") + " " + hits.length) + '</span></div><div class="card-b">' + rows + "</div></div>";
      });
    });
  }
  function wfFlowNodesHtml(wf) {
    const nodes = wf && Array.isArray(wf.executionFlow) ? wf.executionFlow : [];
    if (!nodes.length) {
      return '<div class="small faint">' + esc(t("\u6267\u884C\u9762\u672A\u7ED1\u5B9A\uFF08ConnectorOps \u9AA8\u67B6\u7F3A\u5E2D\uFF09\u2014\u2014workflow apply \u540E\u7684\u843D\u5730 agent \u5F85\u63A5\u7EBF\uFF1B\u5F53\u524D\u8282\u70B9=\u76F8\u4F4D\u94FE\uFF0C\u8282\u70B9\u7EA7\u4F1A\u8BDD\u672A\u5355\u72EC\u8BB0\u5F55\u3002")) + "</div>";
    }
    return '<div class="wf-flow">' + nodes.map((n) => '<div class="wf-flow-node"><div class="wf-flow-name mono">' + esc(String(n.name || "\u2014")) + '</div><div class="small">' + esc(String(n.title || "")) + '</div><div class="small faint">' + esc([n.unit, n.op, n.connector, n.provider].filter(Boolean).join(" \xB7 ")) + "</div></div>").join("") + "</div>";
  }
  function wfRunNodesHtml(run, task, sessionRows = []) {
    const phases = (Array.isArray(run.phases) && run.phases.length ? run.phases : [run.phase || t("\u5355\u8282\u70B9")]).map(String);
    const byTo = /* @__PURE__ */ new Map();
    for (const h of Array.isArray(run.history) ? run.history : []) {
      if (h && h.to) byTo.set(String(h.to), h);
    }
    const cur = String(run.phase || "");
    const rows = phases.map((p) => {
      const h = byTo.get(p);
      const state = p === cur ? t("\u5F53\u524D") : h ? t("\u5DF2\u8D70\u8FC7") : t("\u672A\u5230\u8FBE");
      const anchor = wfSessionAnchor(run, task, sessionRows, h);
      const sessionHtml = p === cur || h ? wfSessionChip(anchor.sid, anchor.label) : '<span class="small faint">' + esc(t("\u65E0\u4F1A\u8BDD\uFF08\u672A\u5230\u8FBE\uFF09")) + "</span>";
      return '<div class="mini-row wf-node-row"><div class="mr-main"><div><b class="mono">' + esc(p) + "</b> " + chip(state, p === cur ? "acc" : h ? "ok" : "mute") + '</div><div class="mr-sub small mono">' + esc(h ? String(h.on || "") : "") + '</div></div><div class="mr-side">' + sessionHtml + (h && h.at ? '<span class="w-ev-time mono">' + esc(relAge(h.at) || "") + "</span>" : "") + "</div></div>";
    }).join("");
    return '<div class="small faint">' + esc(t("\u8282\u70B9\u4F1A\u8BDD\u951A=history.sessionId \u2192 run.sessionId \u2192 Task journal \u53CD\u67E5\uFF08\u7A7A=\u672A\u8BB0\u5F55\uFF0C\u4E0D\u7F16\u9020\uFF1B\u65E7 task.agentSession=taskId \u4E0D\u5F53 session\uFF09\u3002")) + "</div>" + rows;
  }
  function wfTasksHtml(runs, taskIdx, sessionRows = []) {
    const seen = /* @__PURE__ */ new Map();
    for (const r of runs) {
      const key = wfTaskKey(r.involved);
      if (!key) continue;
      if (!seen.has(key)) seen.set(key, { key, task: taskIdx.get(key), runs: [] });
      seen.get(key).runs.push(r);
    }
    if (!seen.size) return "";
    const rows = [...seen.values()].map((x) => {
      const t2 = x.task;
      const title = t2 ? String(t2.title || t2.name || x.key) : x.key;
      const ph = t2 ? String(t2.phase || "") : "";
      const taskSessions = (Array.isArray(sessionRows) ? sessionRows : []).filter((r) => r && r.taskId === x.key && /^session[-_]/i.test(String(r.sessionId || ""))).sort((a, b) => (Number(b.updatedAt) || 0) - (Number(a.updatedAt) || 0));
      const latest = taskSessions[0] || null;
      const direct = /^session[-_]/i.test(String(t2 && t2.agentSession || "")) ? String(t2.agentSession) : "";
      const sid = latest ? String(latest.sessionId) : direct;
      const sessionHtml = wfSessionChip(sid, "\u4EFB\u52A1\u4F1A\u8BDD") + (taskSessions.length > 1 ? '<span class="small faint">' + esc(t("\u5171 {n} \u4E2A\u771F\u5B9E\u4F1A\u8BDD").replace("{n}", String(taskSessions.length))) + "</span>" : "");
      return '<div class="mini-row"><div class="mr-main"><div><b>' + esc(trunc(title, 64)) + "</b> " + (ph ? chip(ph, PH_CLS[ph.toLowerCase()] || "mute") : "") + '</div><div class="mr-sub small mono">Task/' + esc(x.key) + " \xB7 " + esc(t("runs") + " " + x.runs.length) + '</div></div><div class="mr-side">' + sessionHtml + " " + chipGo(t("\u4EFB\u52A1"), "task/" + encodeURIComponent(x.key), "mute", true) + "</div></div>";
    }).join("");
    return '<h2 class="sec-title">' + esc(t("\u4EFB\u52A1")) + '</h2><div class="card">' + rows + "</div>";
  }
  function wfRunsHtml(runs, taskIdx, sessionRows = []) {
    if (!runs.length) return "";
    return '<h2 class="sec-title">' + esc(t("\u8FD0\u884C\u8BB0\u5F55")) + '</h2><div class="card">' + runs.map((r, i) => {
      const no = r.run != null ? r.run : r.no != null ? r.no : i + 1;
      const task = taskIdx.get(wfTaskKey(r.involved));
      const anchor = wfSessionAnchor(r, task, sessionRows);
      return '<details class="ex-block wf-run"><summary class="ex-sum wf-run-sum"><span class="ex-chev">\u25B6</span><b class="mono">run #' + esc(String(no)) + "</b> " + runChip(r) + '<span class="small wf-run-title">' + esc(task ? String(task.title || task.name || "") : String(r.involved || "\u2014")) + '</span><span class="wf-run-side">' + wfSessionChip(anchor.sid, anchor.label) + '<span class="w-ev-time mono">' + esc(relAge(r.startedAt || r.at || r.ts) || "\u2014") + "</span>" + chipGo(t("\u8FD0\u884C\u8BE6\u60C5"), "wf/" + encodeURIComponent(String(r.workflow || "")) + "/r/" + encodeURIComponent(String(no)), "mute", true) + '</span></summary><div class="ex-body">' + wfRunNodesHtml(r, task, sessionRows) + "</div></details>";
    }).join("") + "</div>";
  }
  function wfEditChips(wf, name) {
    const chatFix = '<button type="button" class="btn btn-chip sm" data-wf-chatfix="' + esc(String(name)) + '" title="' + esc(t("\u53D1\u8D77\u5BF9\u8BDD\u4FEE\u6539\uFF08LLM \u6539\u6587\u6863\u2192Apply\u2014\u2014\u4E3B\u901A\u8DEF\uFF09")) + '">' + esc(t("\u5BF9\u8BDD\u4FEE\u6539")) + "</button> ";
    if (wf && String(wf.specSource || "") === "object") {
      const scope = String(wf.scope || "global");
      return chatFix + '<button type="button" class="btn btn-chip sm" data-wf-config-btn title="' + esc(t("\u914D\u7F6E\u8BE6\u60C5\uFF08\u8868\u5355/YAML \u624B\u6539\u2014\u2014\u9ED8\u8BA4\u4E0D\u5C55\u793A\uFF09")) + '">' + esc(t("\u914D\u7F6E\u8BE6\u60C5")) + "</button> " + chipGo(t("\u5BF9\u8C61\u5DE5\u4F5C\u53F0"), objRoute("Workflow", scope, name), "mute", true);
    }
    return chatFix + chipGo(t("\u7F16\u8F91\u5B9A\u4E49\uFF08workflow.md\uFF09"), wfDesignRoute(), "mute", true, t("\u8BBE\u8BA1\u6587\u6863\u5DE5\u4F5C\u53F0")) + " " + chipGo(t("\u76F8\u4F4D\u8868\uFF08scheme\uFF09"), "designdoc/" + encodeURIComponent("core_design") + "/" + encodeURIComponent("core/scheme.yaml"), "mute", true);
  }
  var wfDetailSeq = 0;
  function detailScreen(el, ctx2, name) {
    const stamp = "#" + ++wfDetailSeq;
    el.setAttribute("data-wf-detail-stamp", stamp);
    el.innerHTML = '<div class="view-inner">' + screenHead(t("\u5DE5\u4F5C\u6D41"), String(name), chipGo(t("\u2190 \u5DE5\u4F5C\u6D41"), "workflows", "mute", false)) + '<div data-wf-detail><div class="small faint">' + esc(t("\u5DE5\u4F5C\u6D41\u73B0\u8BFB\u4E2D\u2026\uFF08\u5B9A\u4E49/\u8FD0\u884C/\u4EFB\u52A1\uFF09")) + "</div></div></div>";
    (async () => {
      const [wfsRes, runsRes, tasksRes] = await Promise.all([
        ctx2.api.workflows(),
        ctx2.api.workflowRuns(),
        ctx2.api.tasks()
      ]);
      if (!el.isConnected || el.getAttribute("data-wf-detail-stamp") !== stamp) return;
      const rows = wfApiRows(wfsRes);
      const runsAll = wfApiRuns(runsRes);
      const taskIdx = wfTaskIndex(wfApiTasks(tasksRes));
      const serverAliases = wfsRes && wfsRes.ok && wfsRes.data && wfsRes.data.aliases || {};
      const alias = serverAliases[String(name)] || WF_ALIASES[String(name)];
      let wf = rows.find((x) => String(x && x.name) === String(name));
      if (!wf && alias) wf = rows.find((x) => String(x && x.name) === String(alias));
      const usedName = wf ? String(wf.name) : String(alias || name);
      const runs = runsAll.filter((r) => String(r && r.workflow) === usedName || String(r && r.workflow) === String(name)).sort((a, b) => (Number(b.run) || 0) - (Number(a.run) || 0));
      let sessionRows = [];
      const taskIds = [...new Set(runs.map((r) => wfTaskKey(r.involved)).filter(Boolean))];
      if (taskIds.length && ctx2.api.taskSessions) {
        try {
          const sr = await ctx2.api.taskSessions(taskIds);
          if (sr && sr.ok && Array.isArray(sr.data && sr.data.items)) sessionRows = sr.data.items;
        } catch {
        }
      }
      if (!el.isConnected || el.getAttribute("data-wf-detail-stamp") !== stamp) return;
      const detail = el.querySelector("[data-wf-detail]");
      if (!detail) return;
      if (!rows.length && !runs.length) {
        detail.innerHTML = emptyBox(
          t("\u5DE5\u4F5C\u6D41\u9762\u6682\u4E0D\u53EF\u8BFB"),
          t("GET /v1/workflows \u4E0E /v1/workflows/runs \u5747\u672A\u8FD4\u56DE\u6570\u636E\u2014\u2014") + " " + chipGo(t("\u91CD\u8BD5"), "wf/" + encodeURIComponent(String(name)), "acc", true)
        );
        return;
      }
      if (!wf) {
        const similar = rows.filter((x) => String(x.name || "").toLowerCase().includes(String(name).toLowerCase().split("-")[0])).slice(0, 8);
        detail.innerHTML = emptyBox(t("\u5DE5\u4F5C\u6D41\u4E0D\u5728\u518C"), t("\u6682\u65E0\u6B64\u5DE5\u4F5C\u6D41\uFF08\u53EF\u80FD\u5DF2\u4E0B\u7EBF\u6216\u540D\u79F0\u6709\u8BEF\uFF09\u3002")) + (similar.length ? '<div class="card"><div class="card-h"><span class="h-title">' + esc(t("\u76F8\u8FD1\u5DE5\u4F5C\u6D41")) + '</span></div><div class="card-b">' + similar.map((x) => chipGo(String(x.name), "wf/" + encodeURIComponent(String(x.name)), "mute", true)).join(" ") + "</div></div>" : "");
        return;
      }
      const latest = wf.latest || null;
      const exec = wf.execution || {};
      const aliasNote = String(name) !== usedName ? chip(t("\u8BCD\u5F62\u517C\u5BB9") + "\uFF1A" + String(name) + " \u2192 " + usedName, "mute") : "";
      const kvRows = [
        [t("\u6765\u6E90"), chip(String(wf.specSource || "\u2014"), wf.specSource === "builtin" ? "acc" : wf.specSource === "object" ? "ok" : "mute")],
        [t("\u6267\u884C\u9762"), exec.connector ? chip(String(exec.state || "\u2014"), exec.state === "bound" ? "ok" : "warn") + " " + chipGo(String(exec.connector), "connectors", "mute", true, t("ConnectorOps")) : "\u2014"],
        [t("\u9AA8\u67B6"), exec.skeleton ? chip(t("\u5728\u518C"), "ok") : chip(t("\u7F3A\u5E2D"), "mute")],
        [t("\u8FD0\u884C"), chip(t("\u73AF\u5185 ") + runs.length + " / \u7D2F\u8BA1 " + String(wf.runsTotal != null ? wf.runsTotal : "\u2014"), runs.length ? "acc" : "mute") + (wf.runs24h != null ? " " + chip("24h " + String(wf.runs24h), "mute") : "")],
        [t("\u6700\u65B0"), latest ? runChip(latest) + " " + chip("run #" + String(latest.run), "mono") + " " + esc(relAge(latest.startedAt) || "") : "\u2014"],
        [t("\u751F\u6548"), wf.effective ? chip(String(wf.effective.scope || "") + " \xB7 " + (wf.effective.enabled ? t("\u542F\u7528") : t("\u505C\u7528")), wf.effective.enabled ? "ok" : "mute") : "\u2014"]
      ];
      const runsHtml = wfRunsHtml(runs, taskIdx, sessionRows);
      const tasksHtml = wfTasksHtml(runs, taskIdx, sessionRows);
      detail.innerHTML = '<div class="run-hero"><div class="rh-top">' + (aliasNote ? aliasNote : "") + '</div><div class="rh-title">' + esc(usedName) + " " + chip(String(wf.specSource || "\u2014"), "acc") + " " + chip(String(exec.state || "\u2014"), exec.state === "bound" ? "ok" : "warn") + '</div><div class="rh-meta"><span>' + esc(t("\u5B9A\u4E49/\u8FD0\u884C/\u4EFB\u52A1\u5747\u4E3A\u73B0\u8BFB\u771F\u5B9E\u6570\u636E")) + "</span>" + (latest ? "<span>" + esc(t("\u6700\u8FD1\u8FD0\u884C") + " ") + runChip(latest) + "</span>" : "") + '</div><div class="dh-actions" style="margin-top:10px">' + wfEditChips(wf, usedName) + '</div></div><div class="card"><div class="card-h"><span class="h-title">' + esc(t("\u5B9A\u4E49")) + '</span><span class="h-right">' + // R-UX-10 2026-11-28：YAML 声明=弹窗按需（人不主动看 scheme——fixed 卡退役为按钮）。
      (String(wf.rawYaml || "") ? '<button type="button" class="btn btn-chip sm" data-wf-yaml-btn>' + esc(t("YAML \u58F0\u660E")) + "</button>" : "") + '</span></div><div class="card-b">' + kv(kvRows) + '</div></div><h2 class="sec-title">' + esc(t("\u8282\u70B9\uFF08\u76F8\u4F4D\u94FE\uFF09")) + '</h2><div class="card"><div class="card-b">' + wfNodeGraphHtml(wf, runs) + '</div></div><h2 class="sec-title">' + esc(t("\u6267\u884C\u8282\u70B9\uFF08ConnectorOps \u9AA8\u67B6\uFF09")) + '</h2><div class="card"><div class="card-b">' + wfFlowNodesHtml(wf) + '</div></div><div data-wf-studio-holder style="display:none">' + workflowStudioHtml(wf, usedName) + "</div>" + (tasksHtml || (runs.length ? '<div class="small faint">' + esc(t("\u8FD0\u884C\u672A\u5173\u8054 Task \u5BF9\u8C61\uFF08\u89E6\u53D1\u6E90=Proposal/Template/\u5176\u4ED6\u2014\u2014\u89C1\u8FD0\u884C involved\uFF09\u3002")) + "</div>" : "")) + (runsHtml || '<div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u8BE5 workflow \u6682\u65E0\u8FD0\u884C\u8BB0\u5F55\u2014\u2014")) + (exec.state === "unbound" ? esc(t("\u6267\u884C\u9762\u672A\u7ED1\u5B9A\uFF08ConnectorOps \u9AA8\u67B6\u7F3A\u5E2D\uFF09\uFF0C\u4EFB\u52A1\u4E0D\u4F1A\u62C9\u8D77 run\u3002")) : "") + "</div></div></div>");
      bindWfNodeGraph(el, runs, taskIdx, usedName, sessionRows);
      bindWorkflowStudio(el, ctx2, wf, usedName, () => detailScreen(el, ctx2, usedName));
      const yamlBtn = el.querySelector("[data-wf-yaml-btn]");
      if (yamlBtn) yamlBtn.addEventListener("click", () => {
        htmlModal({ title: t("YAML \u58F0\u660E \xB7 ") + usedName, wide: true, html: yamlTextView(String(wf.rawYaml || "")) });
      });
      const cfgBtn = el.querySelector("[data-wf-config-btn]");
      const studioHolder = el.querySelector("[data-wf-studio-holder]");
      if (cfgBtn && studioHolder) cfgBtn.addEventListener("click", () => {
        const studio = studioHolder.querySelector("[data-wf-studio]");
        if (!studio) {
          toast(t("\u914D\u7F6E\u8BE6\u60C5\u7F3A\u5E2D\uFF08\u8BE5 workflow \u65E0\u53EF\u914D\u7F6E studio\uFF09"), "error");
          return;
        }
        htmlModal({ title: t("\u914D\u7F6E\u8BE6\u60C5 \xB7 ") + usedName, wide: true, node: studio, holder: studioHolder });
      });
      const chatBtn = el.querySelector("[data-wf-chatfix]");
      if (chatBtn) chatBtn.addEventListener("click", () => {
        openDocChat({
          title: t("\u5BF9\u8BDD\u4FEE\u6539 \xB7 Workflow/") + usedName,
          refKey: "obj|Workflow|" + String(wf.scope || "global") + "|" + usedName,
          ref: { kind: "Workflow", scope: String(wf.scope || "global"), name: usedName, rv: Number(wf.resourceVersion) || 0 },
          current: String(wf.rawYaml || ""),
          ctx: ctx2,
          onApplied: () => detailScreen(el, ctx2, usedName)
        });
      });
    })().catch((e) => {
      if (!el.isConnected || el.getAttribute("data-wf-detail-stamp") !== stamp) return;
      const detail = el.querySelector("[data-wf-detail]");
      if (detail) detail.innerHTML = '<div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u5DE5\u4F5C\u6D41\u73B0\u8BFB\u5931\u8D25\uFF1A") + String(e && e.message || e)) + "</div></div></div>";
    });
  }
  function bindDocLinks(scopeEl, ctx2) {
    scopeEl.querySelectorAll("[data-go]").forEach((n) => {
      n.addEventListener("click", () => ctx2.go(n.getAttribute("data-go")));
    });
  }
  function runScreen(el, ctx2, name, no) {
    const runs = ctx2.store.runs().filter((r) => String(r && (r.workflow || r.name || "")) === String(name));
    const run = runs.find((r) => String(r.run != null ? r.run : r.no != null ? r.no : "") === String(no)) || runs[0];
    if (!run) {
      el.innerHTML = '<div class="view-inner"><div class="run-hero"><div class="rh-top">' + chipGo(t("\u2190 " + String(name)), "wf/" + encodeURIComponent(String(name)), "mute", false) + '</div><div class="rh-title">' + esc(String(name)) + ' <span class="rh-run">#' + esc(String(no)) + "</span></div></div>" + emptyBox(t("\u8FD0\u884C\u8BB0\u5F55\u4E0D\u5728\u518C"), t("\u6682\u65E0\u6B64\u8FD0\u884C\u8BB0\u5F55\uFF08\u53EF\u80FD\u5DF2\u5F52\u6863\uFF09\u3002")) + "</div>";
      return;
    }
    let st = String(run.status || "");
    let stCls = st === "suspended" ? "warn" : RUN_CLS[st] || "mute";
    const startedAt = run.startedAt || run.at || run.ts;
    const gateWord = st === "suspended" ? t("\u7B49\u5BA1\u6279") : st === "running" ? t("\u8FD0\u884C\u4E2D") : st === "converged" ? t("\u5DF2\u6536\u675F") : st === "failed" ? t("\u5DF2\u9A73\u56DE") : st;
    const canGate = String(run.phase || "") === "AwaitingApproval" || st === "suspended";
    const gateAnchor = "Workflow/" + String(run.owner || "global") + "/" + String(run.workflow || name);
    el.innerHTML = '<div class="view-inner"><div class="run-hero"><div class="rh-top">' + chipGo(t("\u2190 " + String(name)), "wf/" + encodeURIComponent(String(name)), "mute", false) + '</div><div class="rh-title">' + esc(String(name)) + ' <span class="rh-run">#' + esc(String(run.run != null ? run.run : no)) + "</span>" + chip(st || t("\u672A\u77E5"), stCls) + '</div><div class="rh-meta">' + (startedAt ? "<span>" + esc(t("\u5F00\u59CB\u4E8E")) + " " + esc(relAge(startedAt) || String(startedAt)) + "</span>" : "") + (run.error ? '<span class="rh-err">' + esc(trunc(String(run.error), 120)) + "</span>" : "") + (st === "suspended" ? '<span class="ref-chip" data-go="approvals">' + esc(t("\u53BB\u5BA1\u6279 \u2192")) + "</span>" : "") + "</div>" + (canGate ? '<div class="dh-actions" style="margin-top:10px"><button class="btn btn-primary sm" data-gate-act="grant">' + esc(t("\u6279\u51C6\u7EE7\u7EED")) + '</button> <button class="btn btn-ghost btn-danger sm" data-gate-act="reject">' + esc(t("\u9A73\u56DE\u7EC8\u6B62")) + '</button> <span class="small faint mono">' + esc(t("\u6279\u95F8\u951A\uFF1A")) + esc(gateAnchor) + "</span></div>" : "") + '</div><div class="step-line" data-step-line></div><div class="card" style="margin-bottom:16px"><div class="card-h"><span class="h-title">' + esc(t("\u4F1A\u8BDD")) + '</span></div><div class="card-b" data-run-session>' + wfSessionChip(run.sessionId) + '</div></div><div class="card" style="margin-bottom:16px"><div class="card-h"><span class="h-title">' + esc(t("\u8282\u70B9\u4F1A\u8BDD")) + '</span></div><div class="card-b" data-run-node-sessions><div class="small faint">' + esc(t("\u8282\u70B9\u4F1A\u8BDD\u73B0\u8BFB\u4E2D\u2026")) + '</div></div></div><div class="run-cols"><div class="run-col"><div class="panel-h">' + esc(t("\u5F53\u524D\u76F8\u4F4D")) + '</div><div data-run-now></div></div><div class="run-col"><div class="panel-h">' + esc(t("\u4E8B\u4EF6\u65F6\u95F4\u7EBF")) + '</div><div data-run-tl></div></div></div><div class="foot-note">' + esc(t("\u6D41\u7A0B\u6309\u58F0\u660E\u9636\u6BB5\u63A8\u8FDB\uFF1B" + gateWord + "\u3002")) + "</div></div>";
    const sel = { idx: 0, phases: [] };
    const lineEl = el.querySelector("[data-step-line]");
    const paintLine = () => {
      const phases = Array.isArray(run.phases) && run.phases.length ? run.phases : [run.phase || t("\u5355\u8282\u70B9")];
      sel.phases = phases;
      const curIdx = phases.indexOf(run.phase);
      if (curIdx >= 0) sel.idx = curIdx;
      const visited = new Set((Array.isArray(run.history) ? run.history : []).map((h) => String(h && h.to || "")).filter(Boolean));
      const isTerminal = st === "done" || st === "converged";
      lineEl.innerHTML = phases.map((p, i) => {
        const done = visited.has(p) && (p !== run.phase || isTerminal);
        const active = i === curIdx && !isTerminal;
        const dot = done ? "\u2713" : active ? "" : String(i + 1);
        const conn = i < phases.length - 1 ? '<div class="step-conn' + (done || curIdx >= 0 && i < curIdx ? " fill" : "") + '"></div>' : "";
        return '<div class="step' + (done ? " done" : "") + (active ? " active" : "") + (p === "Failed" ? " failstep" : "") + '" data-step="' + i + '" role="button"><div class="step-dot">' + dot + '</div><div class="step-name">' + esc(p) + "</div></div>" + conn;
      }).join("");
      el.querySelectorAll("[data-step]").forEach((n) => {
        n.addEventListener("click", () => {
          sel.idx = Number(n.getAttribute("data-step"));
          paintNow();
          paintSel();
        });
      });
      paintSel();
    };
    const paintSel = () => {
      el.querySelectorAll("[data-step]").forEach((n, i) => n.classList.toggle("sel", i === sel.idx));
    };
    const nowEl = el.querySelector("[data-run-now]");
    const tlEl = el.querySelector("[data-run-tl]");
    const paintNow = () => {
      const phases = sel.phases.length ? sel.phases : [run.phase || t("\u5355\u8282\u70B9")];
      const ph = phases[sel.idx] || run.phase || t("\u5355\u8282\u70B9");
      const isCur = ph === run.phase;
      const involved = String(run.involved || "");
      const refs = involved.split(",").map((x) => x.trim()).filter((x) => x.split("/").length >= 3).map((x) => {
        const p2 = x.split("/");
        return '<span class="ref-chip" data-go="obj/' + esc(p2[0]) + "/" + esc(p2[1]) + "/" + esc(p2[2]) + '">' + esc(x) + "</span>";
      }).join(" ");
      nowEl.innerHTML = '<div class="now-ph">' + esc(ph) + (isCur ? chip(gateWord || st || t("\u8FDB\u884C\u4E2D"), stCls) : "") + '</div><div class="small faint">' + esc(isCur ? isTerminalRun(st) ? t("\u8BE5\u76F8\u4F4D\u5DF2\u6536\u675F\u2014\u2014\u5355\u6B21\u8FD0\u884C\u7ED3\u675F") : t("\u8FDB\u884C\u4E2D\u2014\u2014\u72B6\u6001\u5B9E\u65F6\u5237\u65B0") : t("\u5386\u53F2\u76F8\u4F4D\u2014\u2014\u6309\u8FD0\u884C\u8F68\u8FF9\u56DE\u653E")) + "</div>" + (refs ? '<div class="now-refs">' + esc(t("\u6D89\u53CA")) + "\uFF1A" + refs + "</div>" : "") + (isCur && st === "suspended" ? '<div class="now-act"><span class="ref-chip" data-go="approvals">' + esc(t("\u53BB\u5904\u7406\u8FD9\u6761\u5BA1\u6279")) + "</span></div>" : "");
      bindDocLinks(nowEl, ctx2);
    };
    const paintTl = () => {
      const trans = Array.isArray(run.history) ? run.history.filter(Boolean) : [];
      const transHtml = trans.length ? '<div class="small faint" style="margin-bottom:6px">' + esc(t("\u9636\u6BB5\u63A8\u8FDB\u8BB0\u5F55\uFF1A")) + '</div><div class="tl" style="margin-bottom:10px">' + trans.slice(0, 20).map((x) => {
        const at = String(x && (x.at || x.ts || x.time) || "");
        const from = String(x && (x.from || x.fromPhase) || "\u2014");
        const to = String(x && (x.to || x.toPhase || x.phase) || "\u2014");
        return '<div class="tl-row"><div class="tl-dot"></div><div class="tl-main"><b>' + esc(from) + " \u2192 " + esc(to) + '</b></div><div class="tl-time mono">' + esc(relAge(at) || (at ? at.slice(0, 16) : "")) + "</div></div>";
      }).join("") + "</div>" : "";
      const evs = ctx2.store.events().filter((e) => {
        const hay = JSON.stringify([e.dedupKey, e.involved, e.type]);
        return hay.includes(String(run.workflow || "")) || hay.includes(String(run.phase || ""));
      }).slice(0, 12);
      tlEl.innerHTML = transHtml + (evs.length ? '<div class="tl">' + evs.map((e) => {
        const eid = e && (e.id != null ? e.id : e.eventId != null ? e.eventId : "");
        const goAttr = eid !== "" ? ' data-go="event/' + encodeURIComponent(String(eid)) + '" role="link" style="cursor:pointer"' : "";
        return '<div class="tl-row"' + goAttr + '><div class="tl-dot"></div><div class="tl-main"><b>' + esc(String(e.type || "")) + '</b><div class="mr-sub small mono">' + esc(trunc(String(e.involved || "\u2014"), 56)) + '</div></div><div class="tl-time">' + esc(relAge(e.ts) || "") + "</div></div>";
      }).join("") + "</div>" : '<div class="tl-empty">' + esc(t("\u6682\u65E0\u76F8\u5173\u4E8B\u4EF6\u2014\u2014\u6D41\u7A0B\u63A8\u8FDB\u65F6\u6B64\u5904\u5B9E\u65F6\u66F4\u65B0\u3002")) + "</div>");
    };
    paintLine();
    paintNow();
    paintTl();
    let nodeSessionRows = [];
    let nodeTask = null;
    const hydrateNodeSessions = async () => {
      const box = el.querySelector("[data-run-node-sessions]");
      if (!box || !el.isConnected) return;
      const key = wfTaskKey(run.involved);
      if (!key) {
        box.innerHTML = '<div class="small faint">' + esc(t("\u8BE5\u8FD0\u884C\u672A\u5173\u8054 Task \u5BF9\u8C61\u2014\u2014\u8282\u70B9\u4F1A\u8BDD\u672A\u8BB0\u5F55\u3002")) + "</div>";
        return;
      }
      try {
        const [tasksRes, sessionsRes] = await Promise.all([
          ctx2.api.tasks(),
          ctx2.api.taskSessions ? ctx2.api.taskSessions([key]) : Promise.resolve({ ok: false })
        ]);
        if (!el.isConnected) return;
        const tasks = wfApiTasks(tasksRes);
        nodeTask = tasks.find((t2) => String(t2.scope || "global") + "/" + String(t2.name || "") === key) || null;
        nodeSessionRows = sessionsRes && sessionsRes.ok && Array.isArray(sessionsRes.data && sessionsRes.data.items) ? sessionsRes.data.items : [];
        box.innerHTML = wfRunNodesHtml(run, nodeTask, nodeSessionRows);
        const topAnchor = wfSessionAnchor(run, nodeTask, nodeSessionRows);
        const sessionBox = el.querySelector("[data-run-session]");
        if (sessionBox) sessionBox.innerHTML = wfSessionChip(topAnchor.sid, topAnchor.label);
      } catch {
        if (el.isConnected) box.innerHTML = '<div class="small faint">' + esc(t("\u8282\u70B9\u4F1A\u8BDD\u8BFB\u53D6\u5931\u8D25\u2014\u2014\u4E0D\u5192\u5145\u5DF2\u8BB0\u5F55\u3002")) + "</div>";
      }
    };
    hydrateNodeSessions();
    const refreshFace = () => ctx2.api.workflowRuns().then((r) => {
      if (!r.ok || !el.isConnected) return;
      const arr = Array.isArray(r.data && r.data.runs) ? r.data.runs : Array.isArray(r.data) ? r.data : [];
      const full = arr.find((x) => String(x && (x.run != null ? x.run : "")) === String(no));
      if (!full) return;
      run.phases = Array.isArray(full.phases) ? full.phases : run.phases;
      run.phase = full.phase || run.phase;
      run.status = full.status || run.status;
      st = String(run.status || "");
      stCls = st === "suspended" ? "warn" : RUN_CLS[st] || "mute";
      run.sessionId = full.sessionId || run.sessionId;
      run.involved = full.involved || run.involved;
      run.transitions = Array.isArray(full.transitions) ? full.transitions : run.transitions;
      run.history = Array.isArray(full.history) ? full.history : run.history;
      const sessionBox = el.querySelector("[data-run-session]");
      if (sessionBox) sessionBox.innerHTML = wfSessionChip(run.sessionId);
      const nodeBox = el.querySelector("[data-run-node-sessions]");
      if (nodeBox && wfTaskKey(run.involved)) nodeBox.innerHTML = wfRunNodesHtml(run, nodeTask, nodeSessionRows);
      paintLine();
      paintNow();
      paintTl();
    });
    refreshFace();
    el.querySelectorAll("[data-gate-act]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const grant = btn.getAttribute("data-gate-act") === "grant";
        const confirmed = await confirmModal({
          title: grant ? t("\u6279\u51C6\u7EE7\u7EED") : t("\u9A73\u56DE\u7EC8\u6B62"),
          body: esc(grant ? t("\u786E\u8BA4\u6279\u51C6\u7EE7\u7EED\uFF1F\u53D1\u5E03 ApprovalGranted\uFF08") + gateAnchor + t("\uFF09\u2014\u2014run \u6CBF\u58F0\u660E\u76F8\u4F4D\u5E8F\u5217\u7EED\u884C\u3002") : t("\u786E\u8BA4\u9A73\u56DE\u7EC8\u6B62\uFF1F\u53D1\u5E03 ApprovalRejected\uFF08") + gateAnchor + t("\uFF09\u2014\u2014run \u5C06\u8FDB\u5165\u7EC8\u6001 Failed\uFF08\u4E0D\u53EF\u6062\u590D\uFF09\u3002")),
          confirmText: grant ? t("\u6279\u51C6") : t("\u9A73\u56DE"),
          danger: !grant
        });
        if (!confirmed) return;
        btn.disabled = true;
        btn.textContent = t("\u53D1\u5E03\u4E2D\u2026");
        const runNo = Number(no) || 0;
        const r = await ctx2.api.eventPub({ type: grant ? "ApprovalGranted" : "ApprovalRejected", involved: gateAnchor, ...runNo > 0 ? { run: runNo } : {} });
        if (!r.ok) {
          btn.disabled = false;
          btn.textContent = grant ? t("\u6279\u51C6\u7EE7\u7EED") : t("\u9A73\u56DE\u7EC8\u6B62");
          toast(esc(t("\u4E8B\u4EF6\u53D1\u5E03\u5931\u8D25\uFF08") + String(r.error || r.status) + t("\uFF09\u2014\u2014run \u4ECD\u6302\u539F\u76F8\u4F4D\uFF0C\u672A\u88C5\u6210\u529F\u3002")), "error");
          return;
        }
        refreshFace();
        ctx2.api.snapshot().then((s) => {
          if (s.ok) ctx2.store.init(s.data);
        });
      });
    });
  }
  function render5(el, ctx2) {
    const p = ctx2.params || {};
    if (p.n != null && p.n !== "") runScreen(el, ctx2, p.name, p.n);
    else if (p.name != null && p.name !== "") detailScreen(el, ctx2, p.name);
    else listScreen(el, ctx2);
  }

  // src/screens/events.js
  var filter = { level: "", q: "" };
  var LEVEL_CLS = { critical: "esc", error: "esc", warn: "warn", warning: "warn", info: "mute", debug: "mute" };
  function clean(x) {
    return String(x == null ? "" : x).replace(/^"+|"+$/g, "");
  }
  function levelChip(e) {
    const lv = String(e && e.level || "info");
    return chip(lv, LEVEL_CLS[lv] || "mute");
  }
  function involvedHtml(e) {
    const s = clean(e && e.involved);
    if (!s) return chip("involved\uFF1A\u2014", "mute");
    return s.split(/[,;]\s*/).filter(Boolean).map((part) => {
      const p = part.trim();
      const route = involvedRoute(p);
      return route ? chipGo(p, route, "mute", true, t("\u67E5\u770B ") + p) : chip(p, "mute");
    }).join(" ");
  }
  function linksHtml(e) {
    const out = [];
    const route = String(e && e.route || "");
    const wf = wfAnchorOfRoute(route);
    if (wf) out.push(chipGo(t("workflow \u951A ") + wf, "wf/" + encodeURIComponent(wf), "acc", true, "route " + route + t(" \u2192 \u5DE5\u4F5C\u6D41")));
    if (e && e.run) out.push(chipGo("run #" + e.run, route && wf ? "wf/" + encodeURIComponent(wf) + "/r/" + encodeURIComponent(String(e.run)) : "workflows", "mute", true, t("\u5173\u8054\u8FD0\u884C")));
    const sess = clean(e && (e.agentSession || e.session));
    if (sess) out.push(chipGo(t("\u4F1A\u8BDD ") + trunc(sess, 18), "session/" + encodeURIComponent(sess), "mute", true, t("\u4F1A\u8BDD\u8BE6\u60C5")));
    const wake = e && (e.wake != null ? e.wake : e.loopId != null ? e.loopId : null);
    if (wake != null) out.push(chipGo(t("\u5FAA\u73AF\u5524\u9192 #") + wake, "welcome", "mute", true, t("\u603B\u89C8\u5FAA\u73AF\u5361\u2014\u2014loop \u5524\u9192\u5173\u8054")));
    return out.join(" ");
  }
  function eventRow(e) {
    const id = e && (e.id != null ? e.id : e.eventId != null ? e.eventId : "");
    const involved = clean(e && e.involved);
    const parts = involved.split(/[,;]\s*/).filter(Boolean);
    const invHtml = parts.length ? parts.map((part) => {
      const p = part.trim();
      const route = involvedRoute(p);
      return route ? chipGo(p, route, "mute", true, t("\u67E5\u770B ") + p) : '<span class="mono">' + esc(p) + "</span>";
    }).join(" ") : '<span class="mono">\u2014</span>';
    const hasClipboard = typeof navigator !== "undefined" && !!(navigator.clipboard && navigator.clipboard.writeText);
    return '<div class="mini-row" data-go="event/' + encodeURIComponent(String(id)) + '"><span class="w-ev-dot" data-level="' + esc(String(e && e.level || "info")) + '"></span><div class="mr-main"><div>' + esc(String(e && e.type || t("\u672A\u547D\u540D\u4E8B\u4EF6"))) + " " + levelChip(e) + '</div><div class="mr-sub mono">' + invHtml + '</div></div><span class="w-ev-time mono">' + esc(relAge(e && (e.ts || e.at || e.time)) || "\u2014") + '</span><button class="btn btn-ghost sm" data-copy-id="' + esc(String(id)) + '"' + (hasClipboard ? "" : ' disabled title="' + esc(t("\u526A\u8D34\u677F\u4E0D\u53EF\u7528")) + '"') + ">" + esc(t("\u590D\u5236 id")) + "</button></div>";
  }
  function listScreen2(el, ctx2) {
    let rows = ctx2.store.events().slice();
    if (filter.level) rows = rows.filter((e) => String(e && e.level || "") === filter.level);
    if (filter.q) {
      const q = filter.q.toLowerCase();
      rows = rows.filter((e) => String(e && (e.type || "") + " " + (e.involved || "") + " " + (e.summary || e.message || "")).toLowerCase().includes(q));
    }
    const fbar = '<div class="filter-bar">' + t('<input class="inp" id="ev-q" placeholder="\u6309\u7C7B\u578B/\u5BF9\u8C61/\u6458\u8981\u8FC7\u6EE4\u2026" value="') + esc(filter.q) + '">' + ["", "error", "warn", "info"].map(
      (lv) => '<span class="fchip' + (filter.level === lv ? " on" : "") + '" data-act="ev-level" data-lv="' + esc(lv) + '">' + esc(lv || t("\u5168\u90E8\u7EA7\u522B")) + "</span>"
    ).join("") + "</div>";
    el.innerHTML = '<div class="view-inner">' + screenHead(t("\u4E8B\u4EF6\u6D41"), t("\u5F52\u6863\u4E8B\u4EF6\uFF08\u5FEB\u7167\u805A\u5408 + patch \u589E\u91CF\uFF1B\u70B9\u51FB\u770B\u5168\u6587\u4E0E\u94FE\u8DEF\uFF09")) + fbar + (rows.length ? '<div class="card">' + rows.slice(0, 300).map(eventRow).join("") + (rows.length > 300 ? t('<div class="small faint" style="padding:8px 14px">\u2026\u5171 ') + rows.length + t(" \u884C\u2014\u2014\u7528\u8FC7\u6EE4\u6536\u7A84\u3002</div>") : "") + "</div>" : emptyBox(t("\u6682\u65E0\u4E8B\u4EF6"), filter.q || filter.level ? t("\u5F53\u524D\u8FC7\u6EE4\u96F6\u547D\u4E2D\u2014\u2014\u653E\u5BBD\u6761\u4EF6\u518D\u8BD5\u3002") : t("\u6682\u65E0\u4E8B\u4EF6\u3002"))) + "</div>";
    const inp = el.querySelector("#ev-q");
    if (inp) {
      inp.addEventListener("input", () => {
        filter.q = inp.value;
        listScreen2(el, ctx2);
      });
      inp.addEventListener("keydown", (ev) => {
        if (ev.key === "Enter") ev.preventDefault();
      });
    }
    el.querySelectorAll('[data-act="ev-level"]').forEach((c) => {
      c.addEventListener("click", () => {
        filter.level = c.dataset.lv || "";
        listScreen2(el, ctx2);
      });
    });
    el.querySelectorAll("[data-copy-id]").forEach((btn) => {
      if (btn.disabled) return;
      btn.addEventListener("click", (ev) => {
        ev.stopPropagation();
        const raw = String(btn.dataset.copyId || "");
        navigator.clipboard.writeText(raw).then(() => {
          btn.textContent = t("\u5DF2\u590D\u5236");
          setTimeout(() => {
            btn.textContent = t("\u590D\u5236 id");
          }, 1200);
        }).catch(() => {
          btn.textContent = t("\u590D\u5236\u5931\u8D25");
          setTimeout(() => {
            btn.textContent = t("\u590D\u5236 id");
          }, 1200);
        });
      });
    });
  }
  async function detailScreen2(el, ctx2, id) {
    const stamp = typeof location !== "undefined" ? location.hash : "";
    el.innerHTML = '<div class="view-inner">' + screenHead(t("\u4E8B\u4EF6\u8BE6\u60C5"), "id=" + id) + '<div class="card"><div class="card-b">' + kv([[t("\u72B6\u6001"), chip(t("\u73B0\u8BFB\u4E2D\u2026"), "mute")]]) + '</div></div><div class="dh-actions" style="margin-top:10px">' + chipGo(t("\u2190 \u56DE\u4E8B\u4EF6\u6D41"), "events", "mute", false) + "</div></div>";
    const r = await ctx2.api.event(id);
    if (!el.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
    const box = el.querySelector(".card .card-b");
    if (!box) return;
    if (!r.ok) {
      box.innerHTML = t('<div class="ob-empty">\u4E8B\u4EF6\u5168\u6587\u6682\u4E0D\u53EF\u8FBE\uFF08') + esc(String(r.error || r.status)) + t("\uFF09\u2014\u2014\u6062\u590D\u540E\u91CD\u8FDB\u5373\u73B0\u8BFB\u3002</div>");
      return;
    }
    const e = r.data && typeof r.data === "object" ? r.data : { raw: r.data };
    const known = ["type", "level", "ts", "at", "time", "involved", "dedupKey", "routedTo", "summary", "message", "route", "run", "agentSession", "session", "wake"];
    const rows = known.filter((k) => e[k] != null && e[k] !== "").map((k) => {
      const v = k === "involved" || k === "summary" || k === "message" || k === "agentSession" || k === "session" ? clean(e[k]) : String(e[k]);
      return [k, k === "ts" || k === "at" || k === "time" ? timeHtml(e[k]) : esc(trunc(v, 200))];
    });
    box.innerHTML = '<div class="ev-head">' + esc(String(e.type || t("\u672A\u547D\u540D\u4E8B\u4EF6"))) + " " + levelChip(e) + " " + linksHtml(e) + "</div>" + (rows.length ? kv(rows) : "") + (e.involved ? '<div class="ev-involved">' + involvedHtml(e) + "</div>" : "") + t('<details class="md-raw"><summary>\u539F\u59CB\u8F7D\u8377\uFF08JSON\uFF09</summary><pre class="mono md-raw-pre">') + esc(JSON.stringify(e, null, 2)) + "</pre></details>";
  }
  function render6(el, ctx2) {
    const id = ctx2.params && ctx2.params.id;
    if (id != null && id !== "") detailScreen2(el, ctx2, String(id));
    else listScreen2(el, ctx2);
  }

  // src/render/meta-header.js
  function frontMetaOf(content, kindHint, scopeHint, nameHint) {
    const { fm } = splitFrontmatter(String(content || ""));
    const tree = fm ? loadYaml(fm) : null;
    const t2 = tree && typeof tree === "object" ? tree : {};
    const meta = t2.metadata && typeof t2.metadata === "object" ? t2.metadata : {};
    const spec = t2.spec && typeof t2.spec === "object" ? t2.spec : {};
    const pick = (...cands) => {
      for (const c of cands) {
        if (c != null && c !== "") return c;
      }
      return null;
    };
    const rawTags = pick(meta.tags, t2.tags, spec.tags) || [];
    const tags = (Array.isArray(rawTags) ? rawTags : []).map((x) => {
      if (x && typeof x === "object" && x.tag) return { tag: String(x.tag), type: String(x.tag_type || x.tagType || "") };
      const s = String(x || "").trim();
      if (!s) return null;
      const i = s.indexOf(":");
      return i > 0 ? { tag: s.slice(i + 1).trim(), type: s.slice(0, i).trim() } : { tag: s, type: "" };
    }).filter(Boolean);
    const rawRefs = pick(t2.ownerRefs, meta.ownerRefs, spec.ownerRefs) || [];
    const ownerRefs = (Array.isArray(rawRefs) ? rawRefs : []).map((r) => {
      if (r && typeof r === "object" && r.kind && r.name) return { kind: String(r.kind), name: String(r.name), scope: String(r.scope || "global") };
      const s = String(r || "").trim();
      const seg = s.split("/").filter(Boolean);
      return seg.length >= 2 ? { kind: seg[0], name: seg.slice(1).join("/"), scope: "global" } : null;
    }).filter(Boolean);
    const impRaw = pick(meta.importance, t2.importance, spec.importance);
    return {
      name: String(pick(t2.name, meta.name, nameHint) || nameHint || ""),
      kind: String(pick(t2.kind, kindHint) || kindHint || ""),
      docType: String(pick(t2.docType, meta.docType, spec.docType) || ""),
      scope: String(pick(t2.scope, meta.scope, scopeHint) || scopeHint || "global"),
      title: String(pick(t2.title, meta.title) || ""),
      tags,
      importance: impRaw != null && Number.isFinite(Number(impRaw)) ? Number(impRaw) : null,
      ownerRefs
    };
  }
  function metaHeaderHtml(meta, history) {
    const m = meta || {};
    const h = history || {};
    const side = h.sidecar && typeof h.sidecar === "object" ? h.sidecar : null;
    const tags = m.tags && m.tags.length ? m.tags : side && Array.isArray(side.tags) ? side.tags.map((x) => x && typeof x === "object" && x.tag ? { tag: String(x.tag), type: String(x.tag_type || "") } : { tag: String(x || ""), type: "" }).filter((x) => x.tag) : [];
    const importance = m.importance != null ? m.importance : side && Number.isFinite(Number(side.importance)) && Number(side.importance) > 0 ? Number(side.importance) : null;
    const parts = [];
    const ident = [];
    if (m.title && m.title !== m.name) ident.push("<b>" + esc(trunc(m.title, 60)) + "</b>");
    ident.push('<b class="mono">' + esc(trunc(String(m.name || ""), 60)) + "</b>");
    if (m.kind) ident.push(kindBadge(m.kind));
    if (m.docType && m.docType !== m.kind) ident.push(chip("docType: " + m.docType, "mute"));
    if (m.scope) ident.push(chip("scope: " + m.scope, "mute"));
    if (h.rv != null) ident.push(chip("r" + String(h.rv), "mute"));
    if (h.head) ident.push(chip(h.headSha ? "HEAD " + String(h.headSha).slice(0, 8) : "HEAD", "mute"));
    parts.push('<div class="mh-ident">' + ident.join(" ") + "</div>");
    const ledger = [];
    if (h.createdAt) ledger.push("<span>" + esc(t("\u521B\u5EFA ")) + '<span class="mono" title="' + esc(String(h.createdAt)) + '">' + esc(String(h.createdAt).replace("T", " ").slice(0, 16)) + (relAge(h.createdAt) ? " \xB7 " + esc(relAge(h.createdAt)) : "") + "</span></span>");
    if (h.lastAt || h.lastBy) {
      const seg = esc(t("\u6700\u8FD1\u63D0\u4EA4 ")) + (h.lastBy ? '<span class="mono">' + esc(trunc(String(h.lastBy), 44)) + "</span>" : "") + (h.lastAt ? '<span class="mono" title="' + esc(String(h.lastAt)) + '">' + (relAge(h.lastAt) ? " \xB7 " + esc(relAge(h.lastAt)) : "") + "</span>" : "");
      ledger.push("<span>" + seg + "</span>");
    }
    if (ledger.length) parts.push('<div class="mh-ledger small faint">' + ledger.join(' <span class="mh-dot">\xB7</span> ') + "</div>");
    if (tags && tags.length || importance != null) {
      const tagHtml = (tags || []).map(
        (x) => chipGo("#" + x.tag, tagRoute(x.tag), "mute", true, x.type ? "tagType: " + x.type : t("\u770B\u540C tag \u6587\u6863"))
      ).join(" ");
      const imp = importance != null ? chip(t("\u91CD\u8981\u5EA6 ") + importance, Number(importance) >= 80 ? "warn" : "mute") : "";
      parts.push('<div class="mh-tags">' + [imp, tagHtml].filter(Boolean).join(" ") + "</div>");
    }
    if (m.ownerRefs && m.ownerRefs.length) {
      const refs = m.ownerRefs.slice(0, 6).map(
        (r) => chipGo(r.kind + "/" + r.name, objRoute(r.kind, r.scope || "global", r.name), "mute", true, t("\u5F52\u5C5E\u5BF9\u8C61"))
      ).join(" ");
      parts.push('<div class="mh-refs small"><span class="faint">' + esc(t("\u5F52\u5C5E\uFF1A")) + "</span> " + refs + "</div>");
    }
    return '<div class="meta-header" data-meta-header>' + parts.join("") + "</div>";
  }
  function paintMetaHeader(slot, ctx2, ref, content, headInfo) {
    if (!slot) return;
    const meta = frontMetaOf(content, ref && ref.kind, ref && ref.scope, ref && ref.name);
    const base = headInfo && headInfo.head ? { head: true, headSha: headInfo.headSha } : { rv: ref ? ref.rv : null };
    if (headInfo && headInfo.sidecar) base.sidecar = headInfo.sidecar;
    slot.innerHTML = metaHeaderHtml(meta, base);
    slot.querySelectorAll("[data-go]").forEach((n) => {
      n.addEventListener("click", (ev) => {
        ev.stopPropagation();
        const c = typeof window !== "undefined" && window.__CC_CONSOLE_CTX__ || null;
        if (c && c.go) c.go(n.getAttribute("data-go"));
      });
    });
    if (!ctx2 || !ctx2.api || typeof ctx2.api.docHistory !== "function" || !ref || !ref.kind || !ref.name) return;
    const stamp = typeof location !== "undefined" ? location.hash : "";
    ctx2.api.docHistory(ref.kind, ref.scope, ref.name).then((r) => {
      if (!slot.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
      const rows = r && r.ok && Array.isArray(r.data && (r.data.rows || r.data.history || r.data.items)) ? r.data.rows || r.data.history || r.data.items : r && r.ok && Array.isArray(r.data) ? r.data : [];
      if (!rows.length) return;
      const first = rows[0];
      const last = rows[rows.length - 1];
      slot.innerHTML = metaHeaderHtml(meta, Object.assign({}, base, {
        createdAt: String(first && (first.at || first.appliedAt) || ""),
        lastAt: String(last && (last.at || last.appliedAt) || ""),
        lastBy: String(last && last.fieldManager || "")
      }));
      slot.querySelectorAll("[data-go]").forEach((n) => {
        n.addEventListener("click", (ev) => {
          ev.stopPropagation();
          const c = typeof window !== "undefined" && window.__CC_CONSOLE_CTX__ || null;
          if (c && c.go) c.go(n.getAttribute("data-go"));
        });
      });
    }).catch(() => {
    });
  }

  // src/screens/objects.js
  function taskHintMentions(sess, low) {
    const s = String(sess || "").trim().toLowerCase();
    if (!s || s.length < 4 || !s.includes("/") || /^session[-_]/i.test(s)) return false;
    return s === low || s.endsWith("/" + low) || s.includes(low);
  }
  function relatedTasksOf(name, tasks) {
    const n = String(name || "").trim();
    if (!n) return [];
    const low = n.toLowerCase();
    const out = [];
    (tasks || []).forEach((t2) => {
      if (!t2) return;
      const spec = t2.spec || {};
      const st = t2.status || t2;
      const id = (t2.scope || "global") + "/" + String(t2.name || "");
      const sess = String(st.agentSession || st.sessionId || "").trim();
      const isSelf = String(t2.name || "") === n;
      const agent = String(spec.agent || t2.agent || "").trim().toLowerCase();
      const owners = (spec.ownerRefs || t2.ownerRefs || []).map((x) => String(x || "").trim().toLowerCase());
      const byAgent = !!agent && agent === low;
      const byOwner = owners.some((r) => r === low || r.endsWith("/" + low));
      const byHint = taskHintMentions(sess, low);
      const hay = [t2.name, spec.title, spec.summary, spec.source, spec.sourceRef, t2.title].map((x) => String(x || "").toLowerCase()).join(" \n");
      const byText = !isSelf && low.length >= 4 && hay.includes(low);
      if (!isSelf && !byAgent && !byOwner && !byHint && !byText) return;
      out.push({
        id,
        self: isSelf,
        title: String(spec.title || t2.title || t2.name || id),
        phase: String(st.phase || ""),
        agent: String(spec.agent || t2.agent || ""),
        session: sess,
        lastAction: String(st.lastAction || "")
      });
    });
    return out.sort((a, b) => (b.self ? 1 : 0) - (a.self ? 1 : 0)).slice(0, 8);
  }
  function relatedRunsOf(name, runs) {
    const n = String(name || "").toLowerCase();
    if (!n || n.length < 4) return [];
    return (runs || []).filter((r) => {
      if (!r) return false;
      const a = String(r.agent || "").trim().toLowerCase();
      if (a && (a === n || a.endsWith("/" + n))) return true;
      return String(r.involved || "").toLowerCase().includes(n);
    }).slice(0, 8).map((r) => ({
      workflow: String(r && (r.workflow || r.name) || ""),
      run: r && (r.run != null ? r.run : r.no != null ? r.no : ""),
      phase: String(r && r.phase || ""),
      status: String(r && r.status || "")
    }));
  }
  function relatedEventsOf(name, events) {
    const n = String(name || "").toLowerCase();
    if (!n || n.length < 4) return [];
    return (events || []).filter((e) => {
      if (!e) return false;
      const inv = String(e.involved || "").toLowerCase();
      const who = String(e.agent || e.owner || "").trim().toLowerCase();
      return inv.includes(n) || !!who && (who === n || who.endsWith("/" + n));
    }).slice(0, 8).map((e) => ({
      type: String(e.type || ""),
      eventId: String(e.eventId || e.id || ""),
      at: String(e.ts || e.at || "")
    }));
  }
  function currentSessionOf(content) {
    let src = String(content || "");
    if (!src.trim()) return "";
    if (src.startsWith("---")) {
      const end = src.indexOf("\n---", 3);
      if (end > 3) src = src.slice(3, end);
    }
    const d = loadYaml(src);
    const cur = d && d.status && d.status.currentSession || d && d.currentSession;
    if (!cur) return "";
    const id = typeof cur === "string" ? cur : cur.id || cur.sessionId || cur.session || "";
    return String(id || "").trim();
  }
  var PHASE_CHIP = {
    running: ["\u6B63\u5728\u505A", "acc"],
    verifying: ["\u9A8C\u8BC1\u4E2D", "acc"],
    queued: ["\u6392\u961F\u4E2D", "mute"],
    ready: ["\u5F85\u542F\u52A8", "mute"],
    pending: ["\u7B49\u5F85\u62C9\u8D77", "mute"],
    blocked: ["\u53D7\u963B", "warn"],
    failed: ["\u5931\u8D25", "esc"],
    done: ["\u5DF2\u5B8C\u6210", "ok"],
    converged: ["\u5DF2\u6536\u655B", "ok"],
    executing: ["\u6267\u884C\u4E2D", "acc"],
    suspended: ["\u6302\u8D77", "warn"],
    canceled: ["\u5DF2\u53D6\u6D88", "mute"]
  };
  function phaseChipText(p) {
    const v = PHASE_CHIP[String(p || "")];
    return v ? chip(v[0], v[1]) : chip(String(p || t("\u72B6\u6001\u672A\u77E5")), "mute");
  }
  function sessionChipHtml(sess) {
    const target = sessionTargetOf(sess);
    if (!target) return "";
    const isTask = target.startsWith("task/");
    return chipGo(
      isTask ? t("\u770B\u4EFB\u52A1") : t("\u770B\u4F1A\u8BDD"),
      target,
      "acc",
      true,
      isTask ? t("\u4EFB\u52A1\u952E\u4F1A\u8BDD\u63D0\u793A\u2014\u2014\u94FE\u4EFB\u52A1\u8BE6\u60C5") : t("\u67E5\u770B\u6267\u884C\u4F1A\u8BDD")
    );
  }
  function relatedPanelHtml(tasks, runs, currentSess, events) {
    if (!tasks.length && !runs.length && !currentSess && !events.length) {
      return t('<div class="ob-empty">\u65E0\u76F8\u5173\u4EFB\u52A1/\u4F1A\u8BDD/\u8FD0\u884C/\u4E8B\u4EF6\uFF08\u5FEB\u7167\u96F6\u547D\u4E2D\u2014\u2014\u4E0D\u7F16\u9020\uFF09\u3002</div>');
    }
    const taskRows = tasks.map(
      (x) => '<div class="mini-row"><div class="mr-main"><div>' + phaseChipText(x.phase) + " <b>" + esc(trunc(x.title, 44)) + "</b> " + (x.agent ? chip(t("\u8D1F\u8D23 ") + x.agent, "mute") : "") + '</div><div class="mr-sub small mono">' + esc(trunc(x.id, 48)) + (x.lastAction ? " \xB7 " + esc(trunc(x.lastAction, 60)) : "") + '</div></div><span class="w-ev-time">' + (x.session ? sessionChipHtml(x.session) : "") + " " + chipGo(t("\u4EFB\u52A1\u8BE6\u60C5"), "task/" + encodeURIComponent(x.id), "mute", true) + "</span></div>"
    ).join("");
    const sessHtml = currentSess ? '<div class="mini-row"><div class="mr-main"><div>' + chip(t("\u5F53\u524D\u4F1A\u8BDD"), "acc") + ' <span class="mono">' + esc(trunc(currentSess, 56)) + '</span></div></div><span class="w-ev-time">' + sessionChipHtml(currentSess) + "</span></div>" : "";
    const runRows = runs.map(
      (x) => '<div class="mini-row"><div class="mr-main"><div>' + phaseChipText(x.phase || x.status) + " <b>" + esc(x.workflow) + "</b> " + chip("run #" + String(x.run != null ? x.run : "?"), "mono") + '</div></div><span class="w-ev-time">' + chipGo(t("\u770B\u8FD0\u884C"), "wf/" + encodeURIComponent(x.workflow) + "/r/" + encodeURIComponent(String(x.run != null ? x.run : "")), "mute", true) + "</span></div>"
    ).join("");
    const eventRows = events.map(
      (x) => '<div class="mini-row"><div class="mr-main"><div>' + chip(t("\u4E8B\u4EF6"), "mute") + " <b>" + esc(trunc(x.type, 44)) + "</b> " + (x.eventId ? chip("#" + String(x.eventId), "mono") : "") + '</div></div><span class="w-ev-time">' + esc(trunc(x.at, 20)) + "</span></div>"
    ).join("");
    return (taskRows ? '<div class="small faint" style="margin-bottom:4px">' + esc(t("\u76F8\u5173\u4EFB\u52A1")) + "</div>" + taskRows : "") + (sessHtml ? '<div class="small faint" style="margin:8px 0 4px">' + esc(t("\u5F53\u524D\u4F1A\u8BDD")) + "</div>" + sessHtml : "") + (runRows ? '<div class="small faint" style="margin:8px 0 4px">' + esc(t("\u76F8\u5173\u5DE5\u4F5C\u6D41\u8FD0\u884C")) + "</div>" + runRows : "") + (eventRows ? '<div class="small faint" style="margin:8px 0 4px">' + esc(t("\u76F8\u5173\u4E8B\u4EF6")) + "</div>" + eventRows : "");
  }
  function tagVocabFromKinds(kinds) {
    const rows = Array.isArray(kinds) ? kinds : [];
    for (const k of rows) {
      const meta = k && k.spec && k.spec.metadata && typeof k.spec.metadata === "object" ? k.spec.metadata : null;
      const t2 = meta && meta.tagTypes;
      if (t2 && typeof t2 === "object" && !Array.isArray(t2)) return t2;
      if (Array.isArray(t2)) return t2.reduce((acc, x) => {
        acc[String(x)] = String(x);
        return acc;
      }, {});
    }
    return {};
  }
  function tagGroups(tags, vocab) {
    const V = vocab || {};
    const UN = t("\u672A\u5206\u7EC4");
    const m = {};
    (tags || []).forEach((t2) => {
      const s = String(t2 || "");
      const colon = s.indexOf(":");
      const prefix = colon > 0 ? s.slice(0, colon) : "";
      const type2 = V[s] || (prefix ? V[prefix] || prefix : "") || UN;
      (m[type2] = m[type2] || []).push(s);
    });
    return Object.keys(m).sort((a, b) => {
      if (a === UN) return 1;
      if (b === UN) return -1;
      return a < b ? -1 : 1;
    }).map((type2) => ({ type: type2, tags: m[type2].sort() }));
  }
  function rulesFor(obj, rules) {
    if (!rules) return { hits: [], degraded: true, reason: t("\u89C4\u5219\u6587\u6863\u65CF\u7D22\u5F15\u6682\u4E0D\u53EF\u8FBE\u2014\u2014\u547D\u4E2D\u5224\u5B9A\u6682\u7F3A\uFF08\u4E0D\u5192\u5145\u65E0\u9002\u7528\uFF09\u3002"), scanned: 0 };
    const o = obj || {};
    const text2 = String(o.content || "");
    const low = text2.toLowerCase();
    const nameLow = String(o.name || "").toLowerCase();
    const scopeLow = String(o.scope || "").toLowerCase();
    const kindLow = String(o.kind || "").toLowerCase();
    const objTags = o.tags || [];
    const out = [];
    (rules || []).forEach((r) => {
      if (!r || !r.path) return;
      const base = String(r.path).replace(/^.*\//, "").replace(/\.md$/, "");
      const rlow = String(r.content || "").toLowerCase();
      const rtags = r.tags || [];
      const why = [];
      const tagHit = objTags.filter((t2) => rtags.includes(t2));
      if (tagHit.length) why.push(t("\u6807\u7B7E\u300C") + tagHit.join("\u3001") + t("\u300D\u540C\u7EC4"));
      if (kindLow && rlow.includes(kindLow)) why.push(t("\u89C4\u5219\u539F\u6587\u63D0\u53CA ") + o.kind);
      if (scopeLow && scopeLow !== "global" && scopeLow !== "builtin" && rlow.includes(scopeLow)) why.push(t("\u89C4\u5219\u539F\u6587\u63D0\u53CA ") + o.scope);
      const nameStem = nameLow.replace(/[-_].*$/, "");
      if (nameStem && nameStem.length >= 3 && rlow.includes(nameStem)) why.push(t("\u89C4\u5219\u539F\u6587\u63D0\u53CA ") + nameStem);
      if (base && low && low.includes(base.toLowerCase())) why.push(t("\u5BF9\u8C61\u6B63\u6587\u5F15\u7528 ") + base);
      if (!why.length) return;
      out.push({ ref: (r.repo || "core_design") + "/" + r.path + "#" + base, name: base, why });
    });
    out.sort((a, b) => b.why.length - a.why.length || (a.name < b.name ? -1 : 1));
    return { hits: out.slice(0, 8), degraded: false, reason: "", scanned: (rules || []).length };
  }
  function sectionCompletion(content, requiredSections) {
    const secs = requiredSections || [];
    if (!secs.length) return [];
    const heads = {};
    String(content || "").split("\n").forEach((line) => {
      const m = /^\s*#{1,6}\s+(.+?)\s*$/.exec(line);
      if (m) heads[m[1].toLowerCase()] = true;
    });
    return secs.map((s) => {
      const key = String(s).toLowerCase();
      return { name: String(s), done: !!heads[key] || String(content || "").toLowerCase().includes(key) };
    });
  }
  function tagsOfDoc(content) {
    const s = String(content || "");
    let yamlSrc = s;
    if (s.startsWith("---")) {
      const end = s.indexOf("\n---", 3);
      if (end > 3) yamlSrc = s.slice(3, end);
    }
    const t2 = loadYaml(yamlSrc);
    if (t2 && typeof t2 === "object") {
      const spec = t2.spec && typeof t2.spec === "object" ? t2.spec : {};
      const topMeta = t2.metadata && typeof t2.metadata === "object" ? t2.metadata : {};
      const specMeta = spec.metadata && typeof spec.metadata === "object" ? spec.metadata : {};
      const raw = topMeta.tags || specMeta.tags || spec.tags || t2.tags;
      if (Array.isArray(raw)) {
        return raw.map((x) => {
          if (x && typeof x === "object" && x.tag) {
            const tt = x.tag_type || x.tagType || "";
            return tt ? String(tt) + ":" + String(x.tag) : String(x.tag);
          }
          return String(x || "").trim();
        }).filter(Boolean);
      }
    }
    const head = s.slice(0, 4e3);
    const flat = head.match(/tags:\s*\[([^\]]*)\]/);
    if (flat) return flat[1].split(",").map((x) => x.trim().replace(/^['"]|['"]$/g, "")).filter(Boolean);
    return [];
  }
  function isMarkdownDoc(kind, name) {
    if (kind === "Goal" || kind === "Rule" || kind === "Runbook" || kind === "DriftReport" || kind === "Proposal" || kind === "Task" || kind === "DesignDoc" || kind === "DocReview" || kind === "BugDoc") return true;
    return /(^|[-_])(README|design|report|notes)([-_.]|$)/i.test(String(name || ""));
  }
  function tagPanel(groups) {
    if (!groups || !groups.length) {
      return t('<div class="ob-empty">\u6682\u65E0\u6807\u7B7E\u2014\u2014\u5BF9\u8C61\u6B63\u6587\u672A\u58F0\u660E spec.metadata.tags\uFF08\u58F0\u660E\u7ECF Apply \u843D\u5E93\u540E\u6B64\u5904\u73B0\u8BFB\uFF09\u3002</div>');
    }
    return groups.map(
      (g) => '<div class="ob-tagrow">' + chip(g.type, "mute") + " " + g.tags.map((t2) => chipGo("#" + t2, tagRoute(t2), "mute", true, t("\u770B\u540C tag \u6587\u6863"))).join(" ") + "</div>"
    ).join("");
  }
  function rulePanel(res) {
    if (res.degraded) return '<div class="ob-empty">' + esc(res.reason) + "</div>";
    if (!res.hits.length) {
      return t('<div class="ob-empty">\u65E0\u9002\u7528\u89C4\u5219\u2014\u2014\u6309 kind/scope/tag/\u6B63\u6587\u5F15\u7528\u56DB\u8DEF\u68C0\u67E5\u96F6\u547D\u4E2D\uFF08\u68C0\u67E5\u8303\u56F4=') + res.scanned + t(" \u4EFD\u89C4\u5219\u6587\u6863\uFF09\u3002</div>");
    }
    return res.hits.map(
      (h, i) => '<details class="ob-rule"><summary><span class="lk mono" data-go="' + esc("designdoc/core_design/" + encodeURIComponent(String(h.ref.replace(/^[^/]+\/([^#]+).*/, "$1") + ".md"))) + '">' + esc(h.ref) + '</span></summary><div class="ob-rule-why small">' + h.why.map((w) => chip(w, "mute")).join(" ") + "</div></details>"
    ).join("");
  }
  function opsPanel(ref, content) {
    let downloadRef = "";
    const m = /download:\s*"?(pkg:\/\/[^"\s}]+)/.exec(String(content || ""));
    if (m) downloadRef = m[1];
    const repoChips = [downloadRef, ref.scope, ref.name].map((x) => repoLinkChip(String(x || ""), String(x || ""))).filter(Boolean).join(" ");
    return repoChips ? '<div class="ob-ctx">' + repoChips + "</div>" : t('<div class="ob-empty">\u65E0\u4ED3\u5E93\u76F4\u8FBE\u58F0\u660E\uFF08connector.download/scope \u8BCD\u5F62\u96F6\u547D\u4E2D\uFF09\u3002</div>');
  }
  function schemePanel(kindReg, content) {
    if (!kindReg) return t('<div class="ob-empty">scheme \u6682\u4E0D\u53EF\u8FBE\uFF08kinds \u805A\u5408\u7F3A\u5E2D\uFF09\u3002</div>');
    const comp = sectionCompletion(content, kindReg.requiredSections || []).map(
      (c) => chip(c.name + (c.done ? " \u2713" : t(" \u7F3A")), c.done ? "ok" : "mute")
    ).join(" ");
    const row = {};
    if (kindReg.name != null) row.name = kindReg.name;
    if (kindReg.requiredSections != null) row.requiredSections = kindReg.requiredSections;
    if (kindReg.spec != null) row.spec = kindReg.spec;
    const yamlText = dumpYaml(row);
    if (!yamlText && !comp) return "";
    return (comp ? '<div class="ob-seccomp" style="margin-bottom:8px">' + comp + "</div>" : "") + (yamlText ? yamlTextView(yamlText, t("kind \u5B9A\u4E49\uFF08YAML \u7EAF\u6587\u672C\uFF09")) : t('<div class="ob-empty">\u8BE5 kind \u65E0\u5B57\u6BB5\u58F0\u660E\uFF08spec \u7F3A\u5E2D\uFF09\u3002</div>'));
  }
  function bindSchemeModal(el, ctx2, kind, content) {
    const btn = el.querySelector("[data-obj-scheme]");
    if (!btn) return;
    btn.addEventListener("click", async () => {
      const paint3 = (kinds) => {
        const k = (kinds || []).find((x) => x && x.name === kind) || null;
        htmlModal({ title: t("scheme \xB7 ") + kind, wide: true, html: schemePanel(k, content || "") });
      };
      const snapKinds = ctx2.store.kinds();
      if (snapKinds.length) {
        paint3(snapKinds);
        return;
      }
      if (ctx2.api && typeof ctx2.api.kinds === "function") {
        try {
          const r = await ctx2.api.kinds();
          paint3(r && r.ok && r.data && Array.isArray(r.data.kinds) ? r.data.kinds : []);
        } catch {
          paint3([]);
        }
      } else paint3([]);
    });
  }
  function contentPanel(kind, name, content) {
    if (!String(content || "").trim()) return t('<div class="ob-empty">\u672C\u6587\u6863\u6682\u65E0\u6B63\u6587\u3002</div>');
    if (isMarkdownDoc(kind, name)) {
      const { body } = splitFrontmatter(content);
      return '<div data-md-render="1">' + renderMarkdown(body) + "</div>" + rawDisclosure(t("\u539F\u6587"), content);
    }
    return yamlTextView(content, t("YAML \u58F0\u660E\uFF08\u7EAF\u6587\u672C\u5168\u6587\uFF09"));
  }
  async function hydrateDoc(el, ctx2, ref) {
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const cached = ctx2.store.cachedDoc(ref.kind, ref.scope, ref.name, ref.rv);
    if (cached) return cached.content;
    const r = await ctx2.api.object(ref.kind, ref.scope, ref.name);
    if (!el.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return null;
    if (!r.ok) return null;
    const data = r.data && typeof r.data === "object" ? r.data : {};
    const content = data.content != null ? String(data.content) : typeof r.data === "string" ? r.data : JSON.stringify(r.data, null, 2);
    ctx2.store.cacheDoc(ref.kind, ref.scope, ref.name, ref.rv, content);
    return content;
  }
  async function hydrateRules(el, ctx2) {
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const ruleRefs = ctx2.store.refs("Rule").slice(0, 12);
    if (!ruleRefs.length) return { hits: [], degraded: false, reason: "", scanned: 0 };
    const rules = [];
    for (const rr of ruleRefs) {
      let content = "";
      const cached = ctx2.store.cachedDoc(rr.kind, rr.scope, rr.name, rr.rv);
      if (cached) content = cached.content;
      else {
        const r = await ctx2.api.object(rr.kind, rr.scope, rr.name);
        if (!el.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return null;
        if (r.ok) {
          const data = r.data && typeof r.data === "object" ? r.data : {};
          content = data.content != null ? String(data.content) : "";
          if (content) ctx2.store.cacheDoc(rr.kind, rr.scope, rr.name, rr.rv, content);
        }
      }
      rules.push({ repo: rr.scope, path: rr.name + ".md", tags: tagsOfDoc(content), content });
    }
    return rules;
  }
  function renderScreen(el, ctx2, ref) {
    const kinds = ctx2.store.kinds();
    const cached = ctx2.store.cachedDoc(ref.kind, ref.scope, ref.name, ref.rv);
    const content = cached ? cached.content : "";
    const tags = tagsOfDoc(content);
    const ruleState = rulesPending.get(ref.kind + "|" + ref.name) || null;
    const related = relatedPanelHtml(
      relatedTasksOf(ref.name, ctx2.store.tasks()),
      relatedRunsOf(ref.name, ctx2.store.runs()),
      currentSessionOf(content),
      // #16 ④ 对象自身 status.currentSession
      relatedEventsOf(ref.name, ctx2.store.events())
    );
    el.innerHTML = '<div class="view-inner">' + screenHead(
      ref.name,
      t("\u5185\u5BB9\u4F18\u5148\uFF1A\u6B63\u6587\u4E0E\u76F8\u5173\u4EFB\u52A1\u4F1A\u8BDD\u76F4\u8FBE\uFF1Bscheme \u7ECF\u6309\u94AE\u6309\u9700\u67E5\u770B"),
      kindBadge(ref.kind) + " " + chip("r" + String(ref.rv || 0), "mute") + ' <button type="button" class="btn btn-chip sm" data-obj-chatfix title="' + esc(t("\u5BF9\u8BDD\u4FEE\u6539\uFF08LLM \u6539\u6587\u6863\u2192Apply\u2014\u2014\u4E3B\u901A\u8DEF\uFF09")) + '">' + esc(t("\u5BF9\u8BDD\u4FEE\u6539")) + '</button> <button type="button" class="btn btn-chip sm" data-obj-scheme title="' + esc(t("scheme \u5B9A\u4E49\uFF08\u5F39\u7A97\u6309\u9700\u67E5\u770B\uFF09")) + '">scheme</button>'
    ) + (pickerHtml(el, ctx2, ref) ? '<div style="margin-bottom:10px">' + pickerHtml(el, ctx2, ref) + (ref.kind === "Provider" ? '<button type="button" class="btn btn-chip sm" data-add-provider style="margin-left:8px" title="' + esc(t("\u65B0\u5EFA Provider\uFF08\u8868\u5355\u624B\u586B=\u964D\u7EA7\u901A\u8DEF\u2014\u2014\u9ED8\u8BA4\u5BF9\u8BDD\u4FEE\u6539\uFF09")) + '">+ ' + esc(t("\u6DFB\u52A0 Provider")) + "</button>" : "") + "</div>" : "") + "<div data-provider-create></div>" + (ref.kind === "Agent" ? '<div data-agent-binding style="margin-bottom:10px"><div class="small faint">' + esc(t("Provider \u7ED1\u5B9A\u73B0\u8BFB\u4E2D\u2026")) + "</div></div>" : "") + t('<section class="card obj-content" style="margin-bottom:12px"><div class="card-h"><span class="h-title">\u5185\u5BB9</span><span class="h-right">') + chip("scope: " + ref.scope, "mute") + " " + chipGo(t("\u7F16\u8F91\u5668"), docRoute(ref.kind, ref.scope, ref.name), "mute", true) + '</span></div><div class="card-b"><div data-meta-header-slot style="margin-bottom:8px">' + (cached ? "" : "") + '</div><div data-panel="content">' + (cached ? contentPanel(ref.kind, ref.name, content) : t('<div class="small faint">\u6B63\u6587\u73B0\u8BFB\u4E2D\u2026</div>')) + "</div></div></section>" + t('<section class="card" style="margin-bottom:12px"><div class="card-h"><span class="h-title">\u76F8\u5173\u4EFB\u52A1\u4E0E\u4F1A\u8BDD</span><span class="h-right">') + chipGo(t("\u6240\u5C5E scope"), "s/" + encodeURIComponent(ref.scope), "mute", true) + '</span></div><div class="card-b">' + related + '</div></section><div class="obj-grid">' + t('<section class="card"><div class="card-h"><span class="h-title">\u6807\u7B7E</span></div><div class="card-b" data-panel="tags">') + tagPanel(tagGroups(tags, tagVocabFromKinds(kinds))) + "</div></section>" + t('<section class="card"><div class="card-h"><span class="h-title">\u9002\u7528\u89C4\u5219</span></div><div class="card-b" data-panel="rules">') + (ruleState ? rulePanel(ruleState) : t('<div class="small faint">\u89C4\u5219\u547D\u4E2D\u5224\u5B9A\u4E2D\u2026</div>')) + '</div></section></div><div class="card" style="margin-top:12px"><div class="card-h"><span class="h-title">' + esc(t("\u64CD\u4F5C")) + '</span></div><div class="card-b">' + opsPanel(ref, cached ? content : "") + "</div></div></div>";
    const picker = el.querySelector("[data-obj-picker]");
    if (picker) {
      picker.addEventListener("change", () => {
        const v = String(picker.value || "");
        if (v) ctx2.go(v);
      });
    }
    bindSchemeModal(el, ctx2, ref.kind, cached ? content : "");
    paintMetaHeader(el.querySelector("[data-meta-header-slot]"), ctx2, ref, cached ? content : "");
    const chatBtn = el.querySelector("[data-obj-chatfix]");
    if (chatBtn) chatBtn.addEventListener("click", async () => {
      let cur = cached ? content : "";
      if (!cur) {
        const r = await ctx2.api.object(ref.kind, ref.scope, ref.name);
        cur = r.ok && r.data && r.data.content != null ? String(r.data.content) : "";
      }
      openDocChat({
        title: t("\u5BF9\u8BDD\u4FEE\u6539 \xB7 ") + ref.kind + "/" + ref.name,
        refKey: "obj|" + ref.kind + "|" + (ref.scope || "global") + "|" + ref.name,
        ref: { kind: ref.kind, scope: ref.scope, name: ref.name, rv: ref.rv },
        current: cur,
        ctx: ctx2,
        onApplied: (modified) => {
          ctx2.store.cacheDoc(ref.kind, ref.scope, ref.name, (ref.rv || 0) + 1, String(modified || ""));
          renderScreen(el, ctx2, ref);
        }
      });
    });
    if (ref.kind === "Provider") hydrateProviderCreate(el, ctx2);
    if (ref.kind === "Agent") hydrateAgentProviderBinding(el, ctx2, ref);
    if (!cached) {
      hydrateDoc(el, ctx2, ref).then((c) => {
        if (c == null || !el.isConnected) return;
        renderScreen(el, ctx2, ref);
      });
    }
    if (!ruleState) {
      hydrateRules(el, ctx2).then((rules) => {
        if (rules == null || !el.isConnected) return;
        const fresh = ctx2.store.cachedDoc(ref.kind, ref.scope, ref.name, ref.rv);
        const res = rulesFor({ kind: ref.kind, scope: ref.scope, name: ref.name, content: fresh ? fresh.content : "", tags }, rules);
        rulesPending.set(ref.kind + "|" + ref.name, res);
        const panel = el.querySelector('[data-panel="rules"]');
        if (panel) panel.innerHTML = rulePanel(res);
      }).catch(() => {
        const panel = el.querySelector('[data-panel="rules"]');
        if (panel) panel.innerHTML = t('<div class="ob-empty">\u89C4\u5219\u547D\u4E2D\u5224\u5B9A\u6682\u4E0D\u53EF\u8BFB\uFF08\u73B0\u8BFB\u5931\u8D25\uFF09\u2014\u2014\u91CD\u8FDB\u672C\u9875\u5373\u91CD\u5224\u3002</div>');
      });
    }
  }
  function pickerHtml(el, ctx2, ref) {
    const objs = ctx2.store.objects();
    if (!objs.length) return "";
    const opts = objs.slice(0, 300).map((o) => {
      const key = o.kind + "/" + (o.scope || "global") + "/" + o.name;
      const cur = ref && o.kind === ref.kind && o.name === ref.name && (o.scope || "global") === (ref.scope || "global");
      return '<option value="' + esc(objRoute(o.kind, o.scope || "global", o.name)) + '"' + (cur ? " selected" : "") + ">" + esc(o.kind + " \xB7 " + o.name) + "</option>";
    }).join("");
    return '<select class="inp" data-obj-picker style="max-width:280px" title="' + esc(t("\u5207\u6362\u5BF9\u8C61")) + '"><option value="">\u2014 ' + esc(t("\u5207\u6362\u5BF9\u8C61\uFF08") + objs.length + t(" \u5728\u518C\uFF09")) + " \u2014</option>" + opts + "</select>";
  }
  var rulesPending = /* @__PURE__ */ new Map();
  function render7(el, ctx2) {
    const p = ctx2.params || {};
    let objs = ctx2.store.objects();
    let ref = null;
    if (p.kind && p.name) {
      ref = objs.find((o) => o.kind === p.kind && o.name === p.name && (!p.scope || p.scope === "global" || o.scope === p.scope)) || null;
    } else if (p.kind) {
      ref = objs.find((o) => o.kind === p.kind) || null;
    } else {
      ref = objs[0] || null;
    }
    if (!ref) {
      const hint = p.kind ? "kind\u300C" + p.kind + "\u300D" : t("\u5FEB\u7167");
      el.innerHTML = '<div class="view-inner">' + screenHead(t("\u5BF9\u8C61\u5DE5\u4F5C\u53F0"), t("\u5185\u5BB9\u4F18\u5148\uFF1A\u6B63\u6587/\u76F8\u5173\u4EFB\u52A1\u4E0E\u4F1A\u8BDD/\u6807\u7B7E/\u89C4\u5219/\u64CD\u4F5C\uFF1Bscheme \u5F39\u7A97\u6309\u9700")) + emptyBox(hint + t("\u6682\u65E0\u5728\u518C\u5BF9\u8C61"), t("\u6682\u65E0\u5728\u518C\u5BF9\u8C61\u3002")) + "</div>";
      return;
    }
    renderScreen(el, ctx2, ref);
  }
  async function hydrateProviderCreate(el, ctx2) {
    const slot = el.querySelector("[data-provider-create]");
    const btn = el.querySelector("[data-add-provider]");
    if (!slot || !btn) return;
    btn.addEventListener("click", () => {
      const form = document.createElement("div");
      form.innerHTML = '<div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:10px">' + prow("name", "openai-2", t("\u5BF9\u8C61\u540D\uFF08\u5982 openai-2\uFF09")) + prow("account", "openai-main", t("\u8D26\u53F7\u6807\u8BC6 (auth.account)")) + prow("secretSrc", "secret:llm/openai-key", t("\u5BC6\u94A5\u5F15\u7528 (auth.secret.source)\u2014\u2014\u96F6\u660E\u6587")) + prow("baseUrl", "https://api.openai.com/v1", t("Base \u5730\u5740 (config.baseUrl)")) + prow("lowModel", "flash-tier", t("Low \u6863\u6A21\u578B")) + prow("stdModel", "standard-tier", t("Standard \u6863\u6A21\u578B")) + prow("highModel", "reasoning-tier", t("High \u6863\u6A21\u578B")) + '</div><div style="margin-top:10px;display:flex;gap:8px;align-items:center"><button type="button" class="btn btn-chip sm" data-prov-create>' + esc(t("\u521B\u5EFA Provider")) + '</button><span class="small faint" data-prov-create-msg></span></div>';
      const createBtn = form.querySelector("[data-prov-create]");
      if (createBtn) createBtn.addEventListener("click", () => createProvider(el, ctx2, form));
      htmlModal({ title: t("\u6DFB\u52A0 Provider\uFF08OpenAI \u517C\u5BB9\uFF09\u2014\u2014\u624B\u586B\u964D\u7EA7\u901A\u8DEF"), wide: true, node: form, holder: slot });
    });
  }
  function prow(key, def2, label) {
    return '<div><div class="small faint">' + esc(label) + '</div><input class="inp mono" data-prov-new="' + key + '" value="' + esc(def2) + '" style="width:100%" /></div>';
  }
  async function createProvider(el, ctx2, slot) {
    const get = (k) => {
      const n = slot.querySelector('[data-prov-new="' + k + '"]');
      return n ? String(n.value || "").trim() : "";
    };
    const msg = slot.querySelector("[data-prov-create-msg]");
    const name = get("name");
    if (!name) {
      if (msg) msg.textContent = t("\u5BF9\u8C61\u540D\u5FC5\u586B");
      return;
    }
    const yaml2 = [
      "apiVersion: core/v1",
      "kind: Provider",
      "metadata:",
      "  name: " + name,
      "  scope: global",
      "spec:",
      "  auth:",
      "    account: " + (get("account") || "openai-main"),
      "    secret:",
      "      source: " + (get("secretSrc") || "secret:llm/openai-key"),
      "  config:",
      "    baseUrl: " + (get("baseUrl") || "https://api.openai.com/v1"),
      "    models:",
      "      low: " + (get("lowModel") || "flash-tier"),
      "      standard: " + (get("stdModel") || "standard-tier"),
      "      high: " + (get("highModel") || "reasoning-tier")
    ].join("\n");
    if (msg) msg.textContent = t("\u6B63\u5728\u521B\u5EFA Provider\u2026");
    const r = await ctx2.api.apply({
      kind: "Provider",
      scope: "global",
      name,
      resourceVersion: 0,
      content: yaml2,
      fieldManager: "console:provider-create",
      reasonRef: "provider-create:" + name
    });
    if (msg) msg.textContent = r.ok ? t("Provider \u5DF2\u521B\u5EFA\u751F\u6548 (rv=" + ((r.data && r.data.rv) != null ? r.data.rv : "?") + ")\u2014\u2014\u5237\u65B0\u540E\u53EF\u5728\u5207\u6362\u5668\u770B\u5230") : t("\u521B\u5EFA\u5931\u8D25\uFF1A") + String(r.error || r.status);
  }
  async function hydrateAgentProviderBinding(el, ctx2, ref) {
    const slot = el.querySelector("[data-agent-binding]");
    if (!slot) return;
    const res = await ctx2.api.object("Agent", ref.scope, ref.name);
    if (!res.ok || !el.isConnected) {
      slot.innerHTML = '<div class="small faint">' + esc(t("Provider \u7ED1\u5B9A\u6682\u4E0D\u53EF\u8BFB")) + "</div>";
      return;
    }
    const raw = String(res.data && res.data.content || "");
    const rv = res.data && res.data.resourceVersion != null ? Number(res.data.resourceVersion) : 0;
    const m = raw.match(/defaultProvider:\s*([^\n#]+)/);
    const cur = m ? m[1].trim().replace(/^["']|["']$/g, "") : "openai-default";
    const provs = (ctx2.store.objects() || []).filter((o) => o.kind === "Provider");
    slot.innerHTML = '<div class="ob-ctx">' + chip(t("\u9ED8\u8BA4 Provider: ") + cur, "ok") + ' <button type="button" class="btn btn-chip sm" data-bind-open>' + esc(t("\u4FEE\u6539\u7ED1\u5B9A")) + '</button> <span class="small faint">rv=' + esc(String(rv)) + "</span></div>";
    const openBtn = slot.querySelector("[data-bind-open]");
    if (!openBtn) return;
    openBtn.addEventListener("click", () => {
      const opts = provs.map((p) => '<option value="' + esc(p.name) + '"' + (p.name === cur ? " selected" : "") + ">" + esc(p.name) + "</option>").join("");
      const form = document.createElement("div");
      form.innerHTML = '<div style="display:flex;gap:10px;align-items:center;flex-wrap:wrap"><span class="small faint">' + esc(t("spec.defaultProvider\uFF08\u7F3A\u7701 openai-default\uFF1B\u4E0D\u540C Agent \u53EF\u7ED1\u4E0D\u540C Provider\uFF09")) + ':</span><select class="inp" data-bind-pick style="max-width:220px"><option value="">' + esc(t("\u2014 \u9009\u62E9\u5728\u518C Provider\uFF08") + provs.length + t(" \u5728\u518C\uFF09")) + " \u2014</option>" + opts + '</select><input class="inp mono" data-bind-free style="width:170px" placeholder="' + esc(t("\u6216\u624B\u586B provider \u540D")) + '" /><button type="button" class="btn btn-chip sm" data-bind-save>' + esc(t("\u4FDD\u5B58\u7ED1\u5B9A")) + '</button><span class="small faint" data-bind-msg></span></div>';
      const pick = form.querySelector("[data-bind-pick]");
      const free = form.querySelector("[data-bind-free]");
      if (pick) pick.addEventListener("change", () => {
        if (free) free.value = String(pick.value || "");
      });
      const save = form.querySelector("[data-bind-save]");
      if (save) save.addEventListener("click", async () => {
        const msg = form.querySelector("[data-bind-msg]");
        const val = String(free && free.value || "").trim().replace(/^["']|["']$/g, "");
        if (!val) {
          if (msg) msg.textContent = t("provider \u540D\u5FC5\u586B");
          return;
        }
        const next = m && m[1] !== void 0 ? raw.replace(/defaultProvider:\s*[^\n#]*/, "defaultProvider: " + val) : raw.replace(/(spec:\n)/, "$1  defaultProvider: " + val + "\n");
        if (msg) msg.textContent = t("\u6B63\u5728\u4FDD\u5B58\u7ED1\u5B9A\u2026");
        const r = await ctx2.api.apply({
          kind: "Agent",
          scope: ref.scope,
          name: ref.name,
          resourceVersion: rv,
          content: next,
          fieldManager: "console:agent-provider-bind",
          reasonRef: "agent-bind:" + ref.name
        });
        if (msg) msg.textContent = r.ok ? t("\u7ED1\u5B9A\u5DF2\u4FDD\u5B58\u751F\u6548 (rv=" + ((r.data && r.data.rv) != null ? r.data.rv : "?") + ")\u2014\u2014\u8BE5 Agent \u540E\u7EED\u8C03\u7528\u6539\u7528\u6B64 Provider") : t("\u4FDD\u5B58\u5931\u8D25\uFF1A") + String(r.error || r.status);
      });
      htmlModal({ title: t("\u4FEE\u6539 Provider \u7ED1\u5B9A \xB7 ") + ref.name, wide: true, node: form, holder: slot });
    });
  }

  // src/screens/editor.js
  function aiFaces() {
    return typeof window !== "undefined" && window.__CC && window.__CC.chatFaces || null;
  }
  var ED = /* @__PURE__ */ new Map();
  var saveTimers = /* @__PURE__ */ new Map();
  var notice = "";
  var screenCtx = null;
  function keyOf(kind, scope, name) {
    return kind + "|" + (scope || "global") + "|" + (name || "");
  }
  function hasDiff(ed) {
    return !!ed && ed.loaded && ed.working !== ed.base;
  }
  function defaultEditorView(kind, name) {
    return isMarkdownDoc(kind, name) ? "render" : "raw";
  }
  function resolveRef(ctx2) {
    const p = ctx2.params || {};
    const objs = ctx2.store.objects();
    if (p.kind && p.name) {
      const hit = objs.find((o) => o.kind === p.kind && o.name === p.name && (!p.scope || p.scope === "global" || o.scope === p.scope));
      if (hit) return { kind: hit.kind, scope: hit.scope, name: hit.name, rv: hit.rv };
      return { kind: p.kind, scope: p.scope || "global", name: p.name, rv: 0 };
    }
    if (p.name) {
      const cands = objs.filter((o) => o.name === p.name);
      if (cands.length === 1) return { kind: cands[0].kind, scope: cands[0].scope, name: cands[0].name, rv: cands[0].rv };
      return null;
    }
    return null;
  }
  async function hydrate(ctx2, key) {
    const { kind, scope, name, rv } = key.ref;
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const faceP = ctx2.api.object(kind, scope, name);
    const row = ctx2.store.drafts().find((d) => d && d.kind === kind && (d.scope || "global") === scope && d.name === name);
    const draftP = row ? ctx2.api.draft(kind, name, scope) : Promise.resolve({ ok: false, status: 404, error: t("\u65E0\u7A3F") });
    const [faceR, draftR] = await Promise.all([faceP, draftP]);
    if ((typeof location !== "undefined" ? location.hash : "") !== stamp) return;
    const prev = ED.get(key.key);
    let base = "";
    let faceRV = 0;
    let faceAbsent = false;
    let faceFail = "";
    if (faceR.ok) {
      const data = faceR.data && typeof faceR.data === "object" ? faceR.data : {};
      base = data.content != null ? String(data.content) : typeof faceR.data === "string" ? faceR.data : JSON.stringify(faceR.data, null, 2);
      faceRV = Number(data.resourceVersion) || rv || 0;
    } else if (faceR.status === 404 && row) {
      faceAbsent = true;
    } else {
      faceFail = String(faceR.error || faceR.status);
    }
    const working = draftR && draftR.ok && draftR.data && draftR.data.content != null ? String(draftR.data.content) : base;
    ED.set(key.key, {
      ref: key.ref,
      working,
      base,
      faceAbsent,
      faceRV,
      faceFail,
      baseRV: draftR && draftR.ok && draftR.data ? Number(draftR.data.baseResourceVersion) || faceRV : faceRV,
      draftFM: draftR && draftR.ok && draftR.data ? String(draftR.data.fieldManager || "") : "",
      view: prev && prev.view || defaultEditorView(kind, name),
      tab: prev && prev.tab || "draft",
      loaded: true
    });
    paintBody(key.key);
    paintBadges(key.key);
    paintRail(key.key, ctx2);
    paintMeta(key.key);
  }
  function bindRetry(keyStr) {
    document.querySelectorAll("[data-ed-retry]").forEach((b) => {
      if (b.getAttribute("data-ed-retry-bound")) return;
      b.setAttribute("data-ed-retry-bound", "1");
      b.addEventListener("click", () => {
        const cur = ED.get(keyStr);
        if (!cur || !screenCtx) return;
        ED.delete(keyStr);
        const body = document.querySelector("[data-ed-body]");
        if (body) body.innerHTML = '<div class="small faint">' + esc(t("\u52A0\u8F7D\u4E2D\u2026")) + "</div>";
        paintMeta(keyStr);
        hydrate(screenCtx, { ref: cur.ref, key: keyStr });
      });
    });
  }
  function paintMeta(keyStr) {
    const slot = document.querySelector("[data-ed-rv]");
    const ed = ED.get(keyStr);
    if (!slot) return;
    if (!ed) slot.innerHTML = esc(t("\u52A0\u8F7D\u4E2D\u2026"));
    else if (ed.faceFail) {
      slot.innerHTML = '<span class="chip esc tiny">' + esc(t("\u6B63\u6587\u62C9\u53D6\u5931\u8D25\uFF08") + ed.faceFail + t("\uFF09")) + '</span> <button class="btn btn-secondary sm" data-ed-retry>' + esc(t("\u91CD\u8BD5")) + "</button>";
      bindRetry(keyStr);
    } else if (ed.faceAbsent) slot.innerHTML = esc(t("\u6587\u6863\u7F3A\u5E2D\uFF08\u5728\u518C\u7A3F\u57FA\u5EA7\u2014\u2014Apply=\u521B\u5EFA\u901A\u8DEF\uFF09"));
    else slot.innerHTML = "r" + ed.faceRV + esc(t("\uFF08\u7248\u672C\u57FA\u51C6\uFF09"));
    paintMetaHeader(
      document.querySelector("[data-ed-meta-header]"),
      screenCtx,
      ed ? ed.ref : null,
      ed ? ed.working : ""
    );
  }
  function paintBody(key) {
    const el = document.querySelector("[data-ed-body]");
    const ed = ED.get(key);
    if (!el || !ed) return;
    if (ed.faceFail) {
      el.innerHTML = '<div class="empty"><div class="e-title">' + esc(t("\u6B63\u6587\u62C9\u53D6\u5931\u8D25\uFF08") + ed.faceFail + t("\uFF09")) + '</div><div class="small faint">' + esc(t("face \u73B0\u8BFB\u672A\u6210\u529F\u2014\u2014\u7248\u672C\u57FA\u51C6\u672A\u77E5\uFF0C\u7F16\u8F91/Apply \u5DF2\u505C\u7528\uFF08\u96F6\u88C5\u5047\uFF09\uFF1B\u70B9\u91CD\u8BD5\u91CD\u8DD1\u6CE8\u6C34\u3002")) + '</div><div class="e-act"><button class="btn btn-secondary sm" data-ed-retry>' + esc(t("\u91CD\u8BD5")) + "</button></div></div>";
      bindRetry(key);
      return;
    }
    el.innerHTML = ed.view === "render" ? isMarkdownDoc(ed.ref.kind, ed.ref.name) ? (function() {
      const { body } = splitFrontmatter(ed.working);
      return '<div data-md-render="1">' + renderMarkdown(body) + "</div>" + rawDisclosure(t("\u539F\u6587"), ed.working);
    })() : '<div data-md-cards="1">' + (yamlToCards(ed.working) || t('<div class="ob-empty">\uFF08\u7A7A\u58F0\u660E\uFF09</div>')) + "</div>" + rawDisclosure("YAML", ed.working) : "";
    const ta = document.querySelector("[data-ed-textarea]");
    if (ta) {
      ta.hidden = ed.view !== "raw";
      if (!ta.hidden && ta.value !== ed.working && document.activeElement !== ta) ta.value = ed.working;
    }
    const renderWrap = document.querySelector('[data-ed-view="render"]');
    if (renderWrap) renderWrap.hidden = ed.view !== "render";
  }
  function paintBadges(key) {
    const ed = ED.get(key);
    if (!ed) return;
    const has = hasDiff(ed);
    const hunks = has ? hunksOf(diffLines2(ed.base, ed.working)) : 0;
    const chipEl = document.querySelector("[data-ed-draft-chip]");
    if (chipEl) {
      chipEl.className = has ? "chip warn tiny" : "chip ok tiny";
      chipEl.textContent = has ? t("\u25CF \u672A Apply\uFF08\u8349\u7A3F ") + hunks + t(" \u5904\u53D8\u66F4\uFF09") : t("\u65E0\u8349\u7A3F");
    }
    const dock = document.querySelector("[data-ed-apply]");
    if (dock) {
      dock.disabled = !has;
      dock.textContent = has ? "Apply\uFF08" + hunks + t(" \u5904\u53D8\u66F4\uFF09") : t("Apply\uFF08\u65E0\u53D8\u66F4\uFF09");
    }
  }
  function docHistoryRowsHtml(rows) {
    if (!Array.isArray(rows) || !rows.length) {
      return '<div class="small faint">' + esc(t("doc_history \u6682\u65E0\u5386\u53F2\u7FFB\u9875\u884C\uFF08\u65B0\u5EFA\u5BF9\u8C61\u9996\u6B21\u843D\u5E93\u524D\u65E0\u65E7\u7248\uFF09\u3002")) + "</div>";
    }
    return rows.slice(0, 40).map((h) => {
      const rev = h && h.rev != null ? String(h.rev) : "?";
      const at = String(h && (h.at || h.appliedAt) || "");
      const fm = String(h && h.fieldManager || "\u2014");
      const bytes = h && h.bytes != null ? String(h.bytes) + " B" : "\u2014";
      return '<div class="mini-row" data-hist-rev="' + esc(rev) + '"><div class="mr-main"><div><b class="mono">r' + esc(rev) + "</b> " + chip(fm, "mute") + " " + chip(bytes, "mute") + '</div><div class="mr-sub small mono">' + esc(at ? at.replace("T", " ").slice(0, 19) : "\u2014") + (at && relAge(at) ? " \xB7 " + esc(relAge(at)) : "") + '</div></div><button class="btn sm" data-rev-peek="' + esc(rev) + '">' + esc(t("\u67E5\u770B\u5FEB\u7167")) + "</button></div>";
    }).join("");
  }
  function paintHistoryTab(box, ed, ctx2) {
    const evRows = ctx2.store.events().filter((e) => e && String(e.involved || "").includes(ed.ref.name)).slice(0, 20);
    const auxHtml = '<details class="ex-block" style="margin-top:12px"><summary class="ex-sum"><span class="ex-chev">\u25B8</span>' + esc(t("\u76F8\u5173\u4E8B\u4EF6\uFF08\u8F85\u52A9\u53C2\u8003\uFF0C\u975E\u7248\u672C\u53F2 \xB7 ")) + evRows.length + '\uFF09</summary><div class="ex-body">' + (evRows.length ? evRows.map((e) => {
      const eid = e && (e.id != null ? e.id : e.eventId != null ? e.eventId : "");
      const goAttr = eid !== "" ? ' data-go="event/' + encodeURIComponent(String(eid)) + '" role="link" style="cursor:pointer"' : "";
      return '<div class="mini-row"' + goAttr + '><div class="mr-main"><div>' + esc(String(e.type || "")) + '</div><div class="mr-sub small mono">' + esc(trunc(String(e.involved || "\u2014"), 50)) + '</div></div><span class="w-ev-time mono">' + esc(String((e.ts || e.at || e.time || "").slice(0, 16))) + "</span></div>";
    }).join("") : '<div class="small faint">' + esc(t("\u5F52\u6863\u7A97\u53E3\u5185\u6682\u65E0\u76F8\u5173\u4E8B\u4EF6\u3002")) + "</div>") + "</div></details>";
    box.innerHTML = t('<div class="card-h"><span class="h-title">\u5386\u53F2</span><span class="h-right">') + chip("doc_history", "acc") + '</span></div><div class="card-b"><div data-ed-hist-list><div class="small faint">' + esc(t("\u9010\u6B21\u6B63\u6587\u53F2\u73B0\u8BFB\u4E2D\u2026\uFF08GET /v1/doc-history\uFF09")) + '</div></div><div data-ed-rev-preview style="margin-top:10px"></div>' + auxHtml + "</div>";
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const fetchHist = ctx2.api && typeof ctx2.api.docHistory === "function" ? ctx2.api.docHistory(ed.ref.kind, ed.ref.scope || "global", ed.ref.name) : Promise.resolve({ ok: false, status: 503, error: "docHistory \u63A5\u53E3\u7F3A\u5E2D" });
    fetchHist.then((r) => {
      if (!box.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
      const listEl = box.querySelector("[data-ed-hist-list]");
      if (!listEl) return;
      if (!r.ok) {
        const msg = r.status === 503 ? t("\u9010\u6B21\u6B63\u6587\u53F2\u672A\u88C5\u914D\uFF08doc_history \u4EC5 sqlite \u540E\u7AEF\u88C5\u914D\uFF1Bgit \u540E\u7AEF\u4E0D\u5192\u5145\u9010\u6B21\u53F2\uFF09") : t("\u5386\u53F2\u6E05\u5355\u6682\u4E0D\u53EF\u8BFB\uFF08") + String(r.error || r.status) + t("\uFF09");
        listEl.innerHTML = '<div class="small faint">' + esc(msg) + "</div>";
        return;
      }
      const rows = r.data && Array.isArray(r.data.history) ? r.data.history : [];
      listEl.innerHTML = docHistoryRowsHtml(rows);
      listEl.querySelectorAll("[data-rev-peek]").forEach((btn) => {
        btn.addEventListener("click", async () => {
          const rv = btn.getAttribute("data-rev-peek");
          const pv = box.querySelector("[data-ed-rev-preview]");
          if (!pv || !rv) return;
          pv.innerHTML = '<div class="small faint">' + esc(t("\u6B63\u53D6 r")) + esc(rv) + esc(t(" \u6B63\u6587\u5FEB\u7167\u2026\uFF08GET /v1/rev-content\uFF09")) + "</div>";
          const rr = ctx2.api && typeof ctx2.api.revContent === "function" ? await ctx2.api.revContent(ed.ref.kind, ed.ref.scope || "global", ed.ref.name, rv) : { ok: false, status: 503, error: "revContent \u63A5\u53E3\u7F3A\u5E2D" };
          if (!pv.isConnected) return;
          if (!rr.ok) {
            pv.innerHTML = '<div class="small faint">' + esc(t("\u5FEB\u7167\u4E0D\u53EF\u8BFB\uFF08")) + esc(String(rr.error || rr.status)) + "\uFF09</div>";
            return;
          }
          const d = rr.data && typeof rr.data === "object" ? rr.data : {};
          const snap = String(d.content || "");
          const src = String(d.source || "doc-history");
          const diff = diffLines2(snap, ed.base || "");
          const st = diffStats(diff);
          pv.innerHTML = '<div class="card"><div class="card-h"><span class="h-title mono">r' + esc(rv) + " " + esc(t("\u5FEB\u7167")) + '</span><span class="h-right">' + chip(src, "mute") + " " + chip("+" + st.add, "ok") + " " + chip("-" + st.del, "esc") + '</span></div><div class="card-b"><div class="small faint" style="margin-bottom:6px">' + esc(t("\u5BF9\u6BD4\u5F53\u524D\u73B0\u503C r")) + esc(String(ed.faceRV || "?")) + "\uFF08" + snap.length + " B\uFF09\uFF1A</div>" + renderDiff(diff, 36) + '<details class="ex-block" style="margin-top:8px"><summary class="ex-sum"><span class="ex-chev">\u25B8</span>' + esc(t("\u67E5\u770B r")) + esc(rv) + esc(t(" \u539F\u6587\u5FEB\u7167")) + '</summary><div class="ex-body"><pre class="mono small" style="max-height:240px;overflow:auto;white-space:pre-wrap">' + esc(trunc(snap, 8e3)) + "</pre></div></details></div></div>";
        });
      });
    }).catch(() => {
      const listEl = box.querySelector("[data-ed-hist-list]");
      if (listEl) listEl.innerHTML = '<div class="small faint">' + esc(t("\u5386\u53F2\u6E05\u5355\u8BFB\u53D6\u5F02\u5E38\u2014\u2014\u4E0D\u5192\u5145\u7A7A\u3002")) + "</div>";
    });
  }
  function paintRail(key, ctx2) {
    const box = document.querySelector("[data-ed-rail]");
    const ed = ED.get(key);
    if (!box || !ed) return;
    if (ed.tab === "ai") {
      const f = aiFaces();
      const canAi = !!(f && typeof f.docFix === "function");
      box.innerHTML = t('<div class="card-h"><span class="h-title">AI \u4FEE\u6587\u6863</span></div><div class="card-b">') + '<div class="small faint">' + esc(t("\u6D41\u7A0B\uFF1A\u5199\u4FEE\u6539\u610F\u56FE \u2192 AI \u51FA\u5168\u6587\u5EFA\u8BAE \u2192 \u5199\u5165\u8349\u7A3F \u2192 \u53F3\u4E0B Apply\uFF08\u7248\u672C\u6821\u9A8C+\u9884\u68C0\u7167\u8D70\u2014\u2014AI \u65E0\u7279\u6743\uFF09\u3002")) + '</div><textarea class="inp" data-ai-prompt rows="3" style="width:100%;margin-top:8px" placeholder="' + esc(t("\u4FEE\u6539\u610F\u56FE\uFF0C\u5982\uFF1A\u8865\u5145 X \u8282\uFF1B\u4FEE\u6B63 Y \u5904\u53E3\u5F84\u2026")) + '"></textarea><button class="btn btn-chip ai-act" data-ai-ask' + (canAi ? "" : " disabled") + ">" + esc(t("\u8BF7\u6C42 AI \u5EFA\u8BAE")) + '</button><div data-ai-out style="margin-top:10px">' + (canAi ? '<div class="small faint">' + esc(t("\u5EFA\u8BAE\u5C06\u73B0\u8BFB\u4E8E\u6B64\u2014\u2014\u5199\u5165\u8349\u7A3F\u540E\u8D70\u65E2\u6709 Apply \u901A\u8DEF\u3002")) + "</div>" : '<div class="small faint">' + esc(t("AI \u901A\u9053\u672A\u6CE8\u5165\uFF08__CC.chatFaces.docFix \u7F3A\u5E2D\uFF09\u2014\u2014\u53EF\u5728\u300C\u539F\u6587\u300D\u89C6\u56FE\u624B\u7F16\uFF0C\u6216\u7ECF\u4F1A\u8BDD\u5C4F\u5BF9\u8BDD\u540E\u4EBA\u5DE5\u8D34\u56DE\u3002"))) + "</div></div>";
      const askBtn = box.querySelector("[data-ai-ask]");
      if (askBtn) askBtn.addEventListener("click", async () => {
        const promptTa = box.querySelector("[data-ai-prompt]");
        const out = box.querySelector("[data-ai-out]");
        if (!promptTa || !out) return;
        const intent = promptTa.value.trim();
        if (!intent) return;
        askBtn.disabled = true;
        out.innerHTML = '<div class="small faint">' + esc(t("AI \u73B0\u8BFB\u4E2D\u2026")) + "</div>";
        try {
          const sug = await f.docFix(intent, { kind: ed.ref.kind, scope: ed.ref.scope, name: ed.ref.name, base: ed.working });
          const text2 = String(sug == null ? "" : sug);
          if (!text2.trim()) throw new Error(t("\u7A7A\u5EFA\u8BAE"));
          out.innerHTML = '<div class="small faint">' + esc(t("AI \u5EFA\u8BAE\u5DF2\u5230\uFF08")) + text2.length + ' B\uFF09</div><pre class="mono small ai-preview">' + esc(trunc(text2, 4e3)) + '</pre><button class="btn btn-secondary sm ai-act" data-ai-apply>' + esc(t("\u5199\u5165\u8349\u7A3F")) + "</button>";
          out.querySelector("[data-ai-apply]").addEventListener("click", () => {
            ed.working = text2;
            paintBody(ed.ref.kind + "/" + (ed.ref.scope || "global") + "/" + ed.ref.name);
            paintBadges(ed.ref.kind + "/" + (ed.ref.scope || "global") + "/" + ed.ref.name);
            out.innerHTML = '<div class="small faint">' + esc(t("\u5DF2\u5199\u5165\u8349\u7A3F\u2014\u2014\u53F3\u4E0B Apply \u843D\u5E93\uFF08\u7248\u672C\u6821\u9A8C\u7167\u8D70\uFF09\u3002")) + "</div>";
          });
        } catch (e) {
          out.innerHTML = '<div class="small faint">' + esc(t("AI \u5EFA\u8BAE\u5931\u8D25\uFF08")) + esc(String(e && e.message || e)) + "\uFF09</div>";
        } finally {
          askBtn.disabled = false;
        }
      });
      return;
    }
    if (ed.tab === "history") {
      paintHistoryTab(box, ed, ctx2);
      return;
    }
    if (!hasDiff(ed)) {
      box.innerHTML = t('<div class="card-h"><span class="h-title">\u8349\u7A3F</span></div><div class="card-b">') + t('<div class="small faint">\u65E0\u8349\u7A3F\u2014\u2014\u7F16\u8F91\u300C\u539F\u6587\u300D\u5373\u7D2F\u79EF\u8349\u7A3F\uFF08\u5199\u7A7F /v1/drafts\uFF0C\u4E0D\u52A8\u6B63\u5F0F\u6587\u6863\u3001\u4E0D\u53D1\u4E8B\u4EF6\uFF09\u3002</div>') + (ed.draftFM ? t('<div class="small faint" style="margin-top:6px">\u7A3F\u6765\u6E90 fieldManager\uFF1A') + esc(ed.draftFM) + "</div>" : "") + "</div>";
      return;
    }
    const lines = diffLines2(ed.base, ed.working);
    const stats = diffStats(lines);
    box.innerHTML = t('<div class="card-h"><span class="h-title">\u8349\u7A3F</span><span class="h-right">') + chip("+" + stats.add, "ok") + " " + chip("-" + stats.del, "esc") + '</span></div><div class="card-b">' + (ed.draftFM ? t('<div class="small faint" style="margin-bottom:6px">\u7A3F\u6765\u6E90 fieldManager\uFF1A') + esc(ed.draftFM) + "</div>" : "") + renderDiff(lines, 40) + t('<div class="small faint" style="margin-top:6px">\u57FA\u51C6=r') + ed.baseRV + t("\uFF08\u843D\u5E93\u671F\u671B\u7248\u672C\u2014\u2014\u7248\u672C\u53D8\u5316\u65F6\u9700\u5237\u65B0\u91CD\u8BD5\uFF09\xB7 \u7A3F ") + ed.working.length + t("B vs \u843D\u5E93\u503C ") + ed.base.length + "B</div></div>";
  }
  function scheduleSave(ctx2, key) {
    const prev = saveTimers.get(key.key);
    if (prev) clearTimeout(prev);
    saveTimers.set(key.key, setTimeout(async () => {
      const ed = ED.get(key.key);
      if (!ed || !hasDiff(ed)) return;
      const r = await ctx2.api.saveDraft({
        kind: ed.ref.kind,
        scope: ed.ref.scope,
        name: ed.ref.name,
        content: ed.working,
        baseResourceVersion: ed.baseRV,
        fieldManager: "console:web"
      });
      if (!r.ok) {
        notice = t("\u8349\u7A3F\u4FDD\u5B58\u5931\u8D25\uFF08") + esc(String(r.error || r.status)) + t("\uFF09\u2014\u2014\u5DE5\u4F5C\u526F\u672C\u4ECD\u5728\u5185\u5B58\uFF0C\u9000\u51FA\u524D\u4E0D\u4E22");
        paintNotice();
      }
    }, 800));
  }
  function paintNotice() {
    const slot = document.querySelector("[data-ed-notice]");
    if (slot) slot.innerHTML = notice ? '<div class="small faint">' + notice + "</div>" : "";
  }
  function editorScreen(el, ctx2, key) {
    const ed = ED.get(key.key);
    const view = ed ? ed.view : "render";
    const bottomHint = isMarkdownDoc(key.ref.kind, key.ref.name) ? t('<div class="small faint">\u9ED8\u8BA4\u5C55\u793A\u6E32\u67D3\u540E\u7684\u6210\u54C1\uFF1B\u70B9\u300C\u539F\u6587\u300D\u67E5\u770B\u4E0E\u7F16\u8F91\u6E90\u6587\u4EF6\u3002\u7F16\u8F91\u5373\u7D2F\u79EF\u8349\u7A3F\uFF08POST /v1/drafts\u2014\u2014\u4E0D\u52A8\u6B63\u5F0F\u6587\u6863\u3001\u4E0D\u53D1\u4E8B\u4EF6\uFF09\uFF1BApply \u552F\u4E00\u5165\u53E3=\u53F3\u680F\u5438\u5E95\u6309\u94AE\u3002</div>') : t('<div class="small faint">YAML \u5BF9\u8C61\u9ED8\u8BA4\u5373\u539F\u6587\uFF08\u7EAF\u6587\u672C\u5168\u6587\uFF09\uFF0C\u65E0\u9700\u70B9\u300C\u539F\u6587\u300D\u67E5\u770B\uFF1B\u4FEE\u6539\u4F18\u5148\u8D70\u300C\u5BF9\u8BDD\u4FEE\u6539\u300D\u4E3B\u901A\u8DEF\uFF08LLM \u6539\u6587\u6863\u2192Apply\uFF09\uFF0C\u624B\u6539\u539F\u6587=\u6700\u540E\u624B\u6BB5\u3002\u7F16\u8F91\u5373\u7D2F\u79EF\u8349\u7A3F\uFF08POST /v1/drafts\u2014\u2014\u4E0D\u52A8\u6B63\u5F0F\u6587\u6863\u3001\u4E0D\u53D1\u4E8B\u4EF6\uFF09\uFF1BApply \u552F\u4E00\u5165\u53E3=\u53F3\u680F\u5438\u5E95\u6309\u94AE\u3002</div>');
    el.innerHTML = '<div class="view-inner"><div class="doc-layout"><div class="doc-main"><div class="doc-headbar"><span class="doc-path mono">' + esc(key.ref.scope + "/" + key.ref.name) + "</span>" + kindBadge(key.ref.kind) + t('<span class="chip ok tiny" data-ed-draft-chip>\u65E0\u8349\u7A3F</span>') + '<div class="md-switch" style="margin-left:auto"><span class="chip click small' + (view === "render" ? " acc" : " mute") + t('" data-ed-view-btn="render">\u6B63\u6587</span>') + '<span class="chip click small' + (view === "raw" ? " acc" : " mute") + t('" data-ed-view-btn="raw">\u539F\u6587</span>') + '</div></div><div class="doc-meta-row">' + chip("scope", "mute") + chipGo(key.ref.scope, "s/" + encodeURIComponent(key.ref.scope), "mute") + '<span class="mono small" data-ed-rv>' + (ed ? ed.faceAbsent ? t("\u6587\u6863\u7F3A\u5E2D\uFF08\u5728\u518C\u7A3F\u57FA\u5EA7\u2014\u2014Apply=\u521B\u5EFA\u901A\u8DEF\uFF09") : "r" + ed.faceRV + t("\uFF08\u7248\u672C\u57FA\u51C6\uFF09") : t("\u52A0\u8F7D\u4E2D\u2026")) + '</span></div><div data-ed-meta-header></div><div data-ed-notice></div><div class="doc-render-wrap" data-ed-view="render" data-ed-body' + (view === "render" ? "" : " hidden") + "></div>" + t('<textarea class="doc-editor" data-ed-textarea spellcheck="false" placeholder="\u539F\u6587\u7F16\u8F91\u2014\u2014\u53D8\u66F4\u5373\u7D2F\u79EF\u8349\u7A3F"') + (view === "raw" ? "" : " hidden") + "></textarea>" + bottomHint + '</div><div class="doc-collab"><div class="collab-tabs"><div class="collab-tab' + ((ed && ed.tab) === "draft" ? " active" : "") + t('" data-ed-tab="draft">\u8349\u7A3F</div>') + '<div class="collab-tab' + ((ed && ed.tab) === "history" ? " active" : "") + t('" data-ed-tab="history">\u5386\u53F2</div>') + '<div class="collab-tab' + ((ed && ed.tab) === "ai" ? " active" : "") + t('" data-ed-tab="ai">AI \u4FEE\u6587\u6863</div>') + '</div><div class="collab-body" data-ed-rail></div>' + t('<div class="apply-dock"><button class="btn btn-primary block" data-ed-apply disabled>Apply\uFF08\u65E0\u53D8\u66F4\uFF09</button></div>') + "</div></div></div>";
    paintNotice();
    paintMeta(key.key);
    if (ed && ed.loaded) {
      paintBody(key.key);
      paintBadges(key.key);
      paintRail(key.key, ctx2);
    } else hydrate(ctx2, key);
    el.querySelectorAll("[data-ed-view-btn]").forEach((b) => {
      b.addEventListener("click", () => {
        const cur = ED.get(key.key);
        if (!cur) return;
        cur.view = b.dataset.edViewBtn;
        el.querySelectorAll("[data-ed-view-btn]").forEach((x) => {
          x.className = "chip click small" + (x.dataset.edViewBtn === cur.view ? " acc" : " mute");
        });
        paintBody(key.key);
      });
    });
    el.querySelectorAll("[data-ed-tab]").forEach((t2) => {
      t2.addEventListener("click", () => {
        const cur = ED.get(key.key);
        if (!cur) return;
        cur.tab = t2.dataset.edTab;
        el.querySelectorAll("[data-ed-tab]").forEach((x) => x.classList.toggle("active", x === t2));
        paintRail(key.key, ctx2);
      });
    });
    const ta = el.querySelector("[data-ed-textarea]");
    if (ta) {
      if (ed && ed.loaded && document.activeElement !== ta) ta.value = ed.working;
      ta.addEventListener("input", () => {
        const cur = ED.get(key.key);
        if (!cur) return;
        cur.working = ta.value;
        paintBadges(key.key);
        paintRail(key.key, ctx2);
        scheduleSave(ctx2, key);
      });
    }
    const dock = el.querySelector("[data-ed-apply]");
    if (dock) {
      dock.addEventListener("click", async () => {
        const cur = ED.get(key.key);
        if (!cur || !hasDiff(cur)) return;
        const stats = diffStats(diffLines2(cur.base, cur.working));
        const okApply = await confirmModal({
          title: t("Apply \u843D\u5E93"),
          body: esc(t("Apply \u843D\u5E93 ") + cur.ref.kind + "/" + cur.ref.name + "\uFF08+" + stats.add + "/-" + stats.del + t("\uFF09\uFF1F\u843D\u5E93\u7248\u672C\u57FA\u51C6 r") + cur.baseRV + "\u3002"),
          confirmText: t("Apply")
        });
        if (!okApply) return;
        dock.disabled = true;
        const r = await ctx2.api.apply({
          kind: cur.ref.kind,
          scope: cur.ref.scope,
          name: cur.ref.name,
          content: cur.working,
          baseResourceVersion: cur.baseRV,
          fieldManager: "console:web",
          reasonRef: "console:web-editor"
        });
        dock.disabled = false;
        if (!r.ok) {
          notice = t("Apply \u5931\u8D25\uFF08") + esc(String(r.error || r.status)) + "\uFF09" + (r.status === 409 ? t("\u2014\u2014\u7248\u672C\u51B2\u7A81\uFF1A\u7A3F\u5DF2\u4FDD\u7559\uFF0C\u8BF7\u5BF9\u9F50\u6700\u65B0\u7248\u672C\u540E\u91CD\u8BD5") : "");
          paintNotice();
          return;
        }
        notice = t("Apply \u6210\u529F\u2014\u2014\u6587\u6863\u5DF2\u843D\u5E93\uFF08rv=") + esc(String(r.data && r.data.rv || "?")) + "\uFF09\u3002";
        cur.base = cur.working;
        paintBadges(key.key);
        paintRail(key.key, ctx2);
        paintNotice();
        ctx2.api.snapshot().then((s) => {
          if (s.ok) ctx2.store.init(s.data);
        });
      });
    }
  }
  function render8(el, ctx2) {
    screenCtx = ctx2;
    const ref = resolveRef(ctx2);
    if (!ref) {
      el.innerHTML = '<div class="view-inner">' + screenHead(t("\u6587\u6863\u7F16\u8F91\u5668"), "") + emptyBox(t("\u5BF9\u8C61\u89E3\u6790\u6B67\u4E49\u6216\u7F3A\u5E2D"), t("\u540C\u540D\u5BF9\u8C61\u8DE8 kind/scope \u4E0D\u552F\u4E00\u2014\u2014\u8BF7\u4ECE\u7CFB\u7EDF\u6811\u8FDB\u5165\u3002"), chipGo(t("\u6253\u5F00\u7CFB\u7EDF\u6811"), "tree", "primary", false)) + "</div>";
      return;
    }
    const key = { ref, key: keyOf(ref.kind, ref.scope, ref.name) };
    editorScreen(el, ctx2, key);
  }

  // src/screens/chat-links.js
  function sessionTasksOf(store2, sessionId) {
    const sid = String(sessionId || "");
    if (!sid) return [];
    return (store2.tasks() || []).filter(
      (t2) => String(t2 && t2.status && (t2.status.agentSession || t2.status.sessionId) || "") === sid
    ).slice(0, 3).map((t2) => ({
      id: String(t2 && t2.scope || "global") + "/" + String(t2 && t2.name || ""),
      title: String(t2 && t2.spec && t2.spec.title || t2 && t2.name || ""),
      phase: String(t2 && t2.status && (t2.status.phase || t2.status) || ""),
      summary: String(t2 && t2.spec && (t2.spec.summary || t2.spec.title) || ""),
      source: String(t2 && t2.spec && (t2.spec.source || t2.spec.sourceRef) || "")
    }));
  }
  function sessBacklinksHtml(store2, sessionId) {
    const sid = String(sessionId || "");
    if (!sid) return "";
    const tasks = sessionTasksOf(store2, sid);
    const evs = (store2.events() || []).filter((e) => {
      const s = String(e && (e.agentSession || e.session) || "");
      return s === sid;
    }).slice(0, 4);
    const taskCards = tasks.map(
      (t2) => '<div class="mini-row"><div class="mr-main"><div>' + (t2.phase ? '<span class="chip acc tiny">' + esc(t2.phase) + "</span> " : "") + "<b>" + esc(trunc(t2.title || t2.id, 44)) + '</b></div><div class="mr-sub small">' + (t2.summary && t2.summary !== t2.title ? esc(trunc(t2.summary, 120)) + "<br>" : "") + (t2.source ? esc(t("\u6765\u6E90\uFF1A")) + '<span class="mono">' + esc(trunc(t2.source, 64)) + "</span> \xB7 " : "") + esc(t("id ")) + '<span class="mono">' + esc(trunc(t2.id, 48)) + '</span></div></div><span class="w-ev-time"><span class="chip acc tiny click" data-go="task/' + encodeURIComponent(t2.id) + '" title="' + esc(t2.id) + t("\u2014\u2014\u4EFB\u52A1\u8BE6\u60C5") + '">' + esc(t("\u4EFB\u52A1\u8BE6\u60C5")) + "</span></span></div>"
    ).join("");
    const evChips = evs.map((e) => '<span class="chip mute tiny click" data-go="event/' + encodeURIComponent(String(e.id ?? "")) + '" title="' + esc(t("\u4E8B\u4EF6\u8BE6\u60C5")) + '">' + esc(trunc(String(e.type || ""), 26)) + "</span>").join(" ");
    if (!taskCards && !evChips) {
      return '<div class="small faint" data-sess-backlinks="1">' + esc(t("\u5FEB\u7167\u53CD\u67E5\uFF08Task.agentSession \u222A \u4E8B\u4EF6 agentSession\uFF09\u65E0\u5728\u518C\u5173\u8054\u2014\u2014\u4F1A\u8BDD\u578B\u72EC\u7ACB\u4F1A\u8BDD\uFF08\u96F6\u7F16\u9020\uFF09\u3002")) + "</div>";
    }
    return '<div data-sess-backlinks="1">' + (taskCards ? '<div class="small faint" style="margin-bottom:4px">' + esc(t("\u6267\u884C\u7684\u4EFB\u52A1\uFF08\u672C\u4F1A\u8BDD\u7531\u4EFB\u52A1\u62C9\u8D77\u2014\u2014\u80CC\u666F\u5982\u4E0B\uFF09")) + "</div>" + taskCards : "") + (evChips ? '<div class="small" style="margin-top:' + (taskCards ? "8" : "0") + 'px"><span class="faint">' + esc(t("\u5173\u8054\u4E8B\u4EF6\uFF1A")) + "</span> " + evChips + "</div>" : "") + "</div>";
  }

  // src/screens/sessionctx.js
  var PANEL = /* @__PURE__ */ new Map();
  function parseTagInput(raw) {
    return String(raw == null ? "" : raw).split(/[\s,，、]+/).map((t2) => t2.trim()).filter(Boolean);
  }
  function capTags(tags, maxTotal = 20, maxPerType = 5) {
    const kept = [];
    const dropped = [];
    const seen = /* @__PURE__ */ new Set();
    const perType = /* @__PURE__ */ new Map();
    for (const raw of Array.isArray(tags) ? tags : []) {
      const key = String(raw || "").trim();
      if (!key || seen.has(key)) continue;
      seen.add(key);
      const c = key.indexOf(":");
      const typ = c > 0 ? key.slice(0, c) : "";
      const n = perType.get(typ) || 0;
      if (kept.length >= maxTotal || n >= maxPerType) {
        dropped.push(key);
        continue;
      }
      perType.set(typ, n + 1);
      kept.push(key);
    }
    return { kept, dropped };
  }
  function normalizeContext(data) {
    const d = data && typeof data === "object" ? data : {};
    const pack = d.pack && typeof d.pack === "object" ? d.pack : {};
    const pol = d.policy && typeof d.policy === "object" ? d.policy : {};
    const sess = d.session && typeof d.session === "object" ? d.session : null;
    const ov = d.override && typeof d.override === "object" ? d.override : null;
    const chunks = (Array.isArray(pack.chunks) ? pack.chunks : []).map((c) => {
      const row = c && typeof c === "object" ? c : {};
      return {
        kind: String(row.kind || ""),
        scope: String(row.scope || ""),
        name: String(row.name || ""),
        score: Number(row.score) || 0,
        why: String(row.why || "")
      };
    });
    const shards = (Array.isArray(d.shards) ? d.shards : []).map((s) => {
      const row = s && typeof s === "object" ? s : {};
      return {
        tagKey: String(row.tagKey || ""),
        docs: Number(row.docs) || 0,
        topDocs: (Array.isArray(row.topDocs) ? row.topDocs : []).map(String),
        stale: !!row.stale
      };
    });
    const stats = d.shardsStats && typeof d.shardsStats === "object" ? {
      total: Number(d.shardsStats.total) || 0,
      stale: Number(d.shardsStats.stale) || 0,
      threshold: Number(d.shardsStats.rebuildThreshold) || 0
    } : null;
    return {
      present: !!(d.sessionId || d.pack || d.policy),
      sessionId: String(d.sessionId || ""),
      session: sess ? {
        state: String(sess.state || ""),
        phase: String(sess.phase || ""),
        lastAction: String(sess.lastAction || ""),
        heartbeats: Number(sess.heartbeats) || 0,
        lastBeat: String(sess.lastBeat || "")
      } : null,
      takeBudget: Number(pack.takeBudget) || 0,
      takeBudgetDefault: Number(pack.takeBudgetDefault) || 0,
      tags: (Array.isArray(pack.tagSig) ? pack.tagSig : []).map((t2) => String((t2 && typeof t2 === "object" ? t2.tagKey : t2) || "")).filter(Boolean),
      tagSigSource: String(pack.tagSigSource || "none"),
      chunks,
      shards,
      stats,
      shardsFacePresent: Array.isArray(d.shards),
      degraded: pack.degraded !== false,
      degradedReason: String(pack.degradedReason || ""),
      policy: {
        shardMaxTokens: Number(pol.shardMaxTokens) || 0,
        minTakeBudget: Number(pol.minTakeBudget) || 0,
        maxTakeBudget: Number(pol.maxTakeBudget) || 0,
        maxTagsPerType: Number(pol.maxTagsPerType) || 5,
        maxTagsTotal: Number(pol.maxTagsTotal) || 20
      },
      override: ov ? {
        revision: Number(ov.revision) || 0,
        takeBudget: Number(ov.takeBudget) || 0,
        updatedAt: String(ov.updatedAt || ""),
        note: String(ov.note || ""),
        by: String(ov.by || "")
      } : null,
      canRevert: !!d.canRevert,
      writeOpen: !(d.writeGate && d.writeGate.open === false)
    };
  }
  function buildSavePayload(model, input) {
    const pol = model && model.policy || {};
    const min = Number(pol.minTakeBudget) || 0;
    const max = Number(pol.maxTakeBudget) || 0;
    const budget = Number(input && input.takeBudget);
    if (!Number.isFinite(budget) || budget < min || budget > max) {
      return { ok: false, error: t("takeBudget \u987B\u4E3A ") + min + "\u2013" + max + t(" \u7684\u6570\u503C\uFF08ContextPolicy \u9884\u7B97\u57DF\uFF09") };
    }
    const { kept, dropped } = capTags(
      parseTagInput(input && input.tagList),
      Number(pol.maxTagsTotal) || 20,
      Number(pol.maxTagsPerType) || 5
    );
    return {
      ok: true,
      dropped,
      payload: {
        takeBudget: Math.round(budget),
        tagSig: kept,
        note: String((input && input.note) == null ? "" : input.note).slice(0, 512),
        by: "console:web",
        baseRevision: model && model.override ? Number(model.override.revision) || 0 : 0
      }
    };
  }
  function sessionRowHtml(m) {
    if (!m.session) {
      return t('<div class="small faint">\u4F1A\u8BDD\u4E0D\u5728\u6D3B\u6027\u8D26\uFF08\u65E0\u5FC3\u8DF3\u884C\uFF09\u2014\u2014\u4ECD\u53EF\u67E5\u770B/\u7F16\u8F91\u672C\u4F1A\u8BDD\u4E0A\u4E0B\u6587\u8986\u76D6\uFF08\u4E0D\u5192\u5145\u5728\u518C\uFF09\u3002</div>');
    }
    const s = m.session;
    const cls = s.state === "suspected-stall" || s.state === "escalated" ? "esc" : s.state ? "acc" : "mute";
    return kv([
      ["state", s.state ? chip(s.state, cls) : '<span class="faint">\u2014</span>'],
      ["phase", esc(s.phase || "\u2014")],
      [t("\u5FC3\u8DF3"), esc(fmtNum(s.heartbeats)) + t(" \u6B21") + (s.lastBeat ? t(" \xB7 \u6700\u8FD1 ") + esc(relAge(s.lastBeat) || t("\u65F6\u95F4\u672A\u77E5")) : "")],
      [t("\u52A8\u4F5C"), esc(trunc(s.lastAction || "\u2014", 80))]
    ]);
  }
  function chunksHtml(m) {
    if (!m.chunks.length) {
      return t('<div class="small faint">\u65E0\u7D22\u5F15\u5019\u9009\uFF08tagSig \u672A\u88C5\u914D\u6216\u7D22\u5F15\u96F6\u547D\u4E2D\u2014\u2014\u4E0D\u5192\u5145\u88C5\u914D\u5305\u5185\u5BB9\uFF09\u3002</div>');
    }
    return m.chunks.map((c) => {
      const key = [c.kind, c.scope, c.name].filter(Boolean).join("/");
      const go = c.kind && c.name ? objRoute(c.kind, c.scope || "global", c.name) : "";
      const titleHtml = go ? '<span class="lk mono" data-go="' + esc(go) + '">' + esc(trunc(key, 64)) + "</span>" : '<b class="mono">' + esc(trunc(key, 64)) + "</b>";
      return '<div class="mini-row"' + (go ? ' data-go="' + esc(go) + '"' : "") + '><div class="mr-main"><div>' + titleHtml + " " + chip(String(Math.round(c.score)), c.score >= 60 ? "ok" : "mute") + "</div>" + (c.why ? '<div class="mr-sub small">' + esc(trunc(c.why, 90)) + "</div>" : "") + "</div></div>";
    }).join("");
  }
  function topDocRoute(d) {
    const s = String(d || "").trim();
    if (!s) return "";
    const p = s.split("/");
    if (p.length >= 3) return objRoute(p[0], p[1], p.slice(2).join("/"));
    if (p.length === 1) return docRoute("Doc", "global", s);
    return "";
  }
  function shardsHtml(m) {
    if (!m.shardsFacePresent) {
      return t('<div class="small faint">\u5206\u7247\u7F13\u5B58\u9762\u7F3A\u5E2D\uFF08contextidx \u672A\u66B4\u9732 Shards\u2014\u2014\u8BDA\u5B9E\u964D\u7EA7\uFF0C\u4E0D\u5192\u5145\u7A7A\u7F13\u5B58\uFF09\u3002</div>');
    }
    if (!m.shards.length) {
      return t('<div class="small faint">\u65E0\u751F\u6548 tag \u5206\u7247\uFF08tagSig \u672A\u88C5\u914D\u2014\u2014\u7F16\u8F91\u5206\u53C9\u540E\u9010 tagKey \u6295\u5F71 TagCtxShard\uFF09\u3002</div>');
    }
    return m.shards.map((s) => {
      const topHtml = s.topDocs.length ? s.topDocs.slice(0, 4).map((d) => {
        const go = topDocRoute(d);
        return go ? '<span class="lk mono small" data-go="' + esc(go) + '">' + esc(trunc(String(d), 40)) + "</span>" : '<span class="mono small faint">' + esc(trunc(String(d), 40)) + "</span>";
      }).join(" \xB7 ") : "";
      return '<div class="mini-row"><div class="mr-main"><div>' + chipGo(s.tagKey, tagRoute(s.tagKey), "mute", true, t("\u770B\u540C tag \u6587\u6863")) + " " + (s.stale ? chip("stale", "warn") : chip("fresh", "ok")) + "</div>" + t('<div class="mr-sub small">\u76F8\u5173\u5BF9\u8C61 ') + esc(fmtNum(s.docs)) + t(" \u4E2A") + (topHtml ? " \xB7 " + topHtml : "") + "</div></div></div>";
    }).join("");
  }
  function formHtml(m) {
    if (!m.writeOpen) {
      return t('<div class="small faint">\u5199\u9762\u5173\u65AD\uFF08--allow-apply \u672A\u5F00\u542F\uFF09\u2014\u2014\u7F16\u8F91\u5206\u53C9\u4E0D\u53EF\u7528\uFF08\u4E0E\u5FC3\u8DF3\u6CE8\u518C\u9762\u540C\u5F00\u5173\uFF0C\u4E0D\u5192\u5145\u53EF\u7F16\u8F91\uFF09\u3002</div>');
    }
    return '<div class="filter-bar" data-ctx-form><input class="inp" data-ctx-budget type="number" min="' + m.policy.minTakeBudget + '" max="' + m.policy.maxTakeBudget + '" step="500" value="' + m.takeBudget + t('" title="\u63D0\u53D6\u9884\u7B97\uFF08token \u6570\uFF09">') + t('<input class="inp" data-ctx-tags placeholder="tag \u96C6\uFF1AtagType:tag \u9017\u53F7/\u7A7A\u683C\u5206\u9694\uFF08\u6BCF\u7C7B\u2264') + m.policy.maxTagsPerType + t(" \u603B\u2264") + m.policy.maxTagsTotal + '\uFF09">' + t('<input class="inp" data-ctx-note placeholder="\u5206\u53C9\u5907\u6CE8\uFF08\u53EF\u9009\uFF09">') + t('<button class="btn btn-secondary sm" data-ctx-save>\u4FDD\u5B58\u5206\u53C9</button>') + '<button class="btn btn-ghost sm" data-ctx-revert' + (m.canRevert ? "" : " disabled") + t(">\u64A4\u9500\u4E0A\u6B21</button>") + "</div>";
  }
  function renderSessionCtx(m) {
    const headChips = (m.degraded ? chip(t("\u7D22\u5F15\u5019\u9009\u6295\u5F71"), "warn") : "") + (m.writeOpen ? chip(t("\u5199\u9762\u5F00"), "ok") : chip(t("\u5199\u9762\u5173\u65AD"), "esc")) + (m.canRevert ? chip(t("\u53EF\u64A4\u9500"), "acc") : "");
    const ovLine = m.override ? t('<div class="small faint" style="margin-top:6px">\u73B0\u884C\u8986\u76D6\uFF1Arevision ') + esc(fmtNum(m.override.revision)) + " \xB7 takeBudget " + esc(fmtNum(m.override.takeBudget)) + " \xB7 by " + esc(m.override.by || "\u2014") + (m.override.note ? t(" \xB7 \u5907\u6CE8 ") + esc(trunc(m.override.note, 60)) : "") + (m.override.updatedAt ? " \xB7 " + esc(relAge(m.override.updatedAt) || "") : "") + "</div>" : t('<div class="small faint" style="margin-top:6px">\u65E0\u8986\u76D6\uFF08\u751F\u6548\u503C=ContextPolicy \u7F3A\u7701\u2014\u2014PUT \u540E\u6B64\u5904\u73B0\u8BFB\u8986\u76D6\u884C\uFF09\u3002</div>');
    return t('<div class="card-h"><span class="h-title">\u4E0A\u4E0B\u6587</span><span class="h-right">') + headChips + '</span></div><div class="card-b">' + sessionRowHtml(m) + '<div style="margin-top:10px">' + kv([
      ["takeBudget", '<b class="mono">' + esc(fmtNum(m.takeBudget)) + t('</b> <span class="small faint">\uFF08\u7F3A\u7701 ') + esc(fmtNum(m.takeBudgetDefault)) + t(" \xB7 \u57DF ") + m.policy.minTakeBudget + "\u2013" + m.policy.maxTakeBudget + "\uFF09</span>"],
      ["tagSig", m.tags.length ? m.tags.map((t2) => chipGo(t2, tagRoute(t2), "mute", true, t("\u770B\u540C tag \u6587\u6863"))).join(" ") : t('<span class="faint small">\u672A\u88C5\u914D\uFF08tagSigSource=') + esc(m.tagSigSource) + "\uFF09</span>"],
      [t("\u5206\u7247\u7F13\u5B58"), m.stats ? t("\u5171 ") + esc(fmtNum(m.stats.total)) + t(" \u7247 \xB7 stale ") + esc(fmtNum(m.stats.stale)) + t(" \xB7 \u91CD\u5EFA\u9608\u503C ") + esc(fmtNum(m.stats.threshold)) : '<span class="faint">\u2014</span>'],
      ["shardMaxTokens", '<span class="mono">' + esc(fmtNum(m.policy.shardMaxTokens)) + "</span>"]
    ]) + "</div>" + ovLine + '<div data-ctx-notice></div><div style="margin:10px 0 4px">' + formHtml(m) + "</div>" + t('<h3 class="sec-title" style="margin:10px 0 6px">chunks\uFF08\u7D22\u5F15 TopK \u5019\u9009\uFF09</h3>') + chunksHtml(m) + t('<h3 class="sec-title" style="margin:10px 0 6px">\u76F8\u5173 Tag \u5206\u7247\uFF08TagCtxShard\uFF09</h3>') + shardsHtml(m) + (m.degraded && m.degradedReason ? '<div class="small faint" style="margin-top:8px">' + esc(m.degradedReason) + "</div>" : "") + "</div>";
  }
  function focusIn(box) {
    return !!(document.activeElement && box.contains(document.activeElement));
  }
  function paint(box, ctx2, id, model) {
    PANEL.set(id, { model });
    box.innerHTML = '<div class="card" data-ctx-card>' + renderSessionCtx(model) + "</div>";
    wireSessionCtx(box, ctx2, id);
  }
  async function refresh(ctx2, id, box) {
    const r = await ctx2.api.sessionContext(id);
    if (!r.ok || focusIn(box)) return;
    paint(box, ctx2, id, normalizeContext(r.data));
  }
  function wireSessionCtx(box, ctx2, id) {
    const q = (sel) => box.querySelector(sel);
    const notice2 = (html3) => {
      const slot = q("[data-ctx-notice]");
      if (slot) slot.innerHTML = html3;
    };
    const saveBtn = q("[data-ctx-save]");
    if (saveBtn) saveBtn.addEventListener("click", async () => {
      const cur = (PANEL.get(id) || {}).model;
      if (!cur) return;
      const built = buildSavePayload(cur, {
        takeBudget: q("[data-ctx-budget]").value,
        tagList: q("[data-ctx-tags]").value,
        note: q("[data-ctx-note]").value
      });
      if (!built.ok) {
        notice2('<div class="small faint">' + esc(built.error) + "</div>");
        return;
      }
      saveBtn.disabled = true;
      const r = await ctx2.api.saveSessionContext(id, built.payload);
      saveBtn.disabled = false;
      if (!r.ok) {
        notice2(t('<div class="small faint">\u4FDD\u5B58\u5931\u8D25\uFF08') + esc(String(r.error || r.status)) + "\uFF09" + (r.status === 409 ? t("\u2014\u2014\u7248\u672C\u51B2\u7A81\uFF1A\u8BF7\u5237\u65B0\u9762\u677F\u5BF9\u9F50\u6700\u65B0\u7248\u672C\u540E\u91CD\u8BD5\uFF08\u539F\u7A3F\u4E0D\u8986\u76D6\uFF09") : "") + "</div>");
        return;
      }
      notice2(t('<div class="small faint">\u5DF2\u4FDD\u5B58\uFF08revision ') + esc(String(r.data && r.data.override && r.data.override.revision || "?")) + "\uFF09" + (built.dropped.length ? t("\uFF1B\u8D85\u9650 tag \u6536\u655B ") + built.dropped.length + t(" \u4E2A\uFF1A") + esc(trunc(built.dropped.join(", "), 80)) : "") + "\u3002</div>");
      refresh(ctx2, id, box);
    });
    const revertBtn = q("[data-ctx-revert]");
    if (revertBtn) revertBtn.addEventListener("click", async () => {
      const okRevert = await confirmModal({
        title: t("\u64A4\u9500\u4E0A\u4E0B\u6587\u5206\u53C9"),
        body: esc(t("\u64A4\u9500\u6700\u8FD1\u4E00\u6B21\u4E0A\u4E0B\u6587\u5206\u53C9\uFF08\u6062\u590D\u670D\u52A1\u7AEF\u5199\u524D\u5FEB\u7167\uFF09\uFF1F")),
        confirmText: t("\u64A4\u9500"),
        danger: true
      });
      if (!okRevert) return;
      revertBtn.disabled = true;
      const r = await ctx2.api.revertSessionContext(id);
      revertBtn.disabled = false;
      if (!r.ok) {
        notice2(t('<div class="small faint">\u64A4\u9500\u5931\u8D25\uFF08') + esc(String(r.error || r.status)) + t("\uFF09\u2014\u2014\u6808\u7A7A\u4E0D\u53EF\u64A4\uFF08\u4E0D\u5192\u5145\u64A4\u9500\u6210\u529F\uFF09\u3002</div>"));
        return;
      }
      notice2(t('<div class="small faint">\u5DF2\u64A4\u9500\u2014\u2014\u5199\u524D\u5FEB\u7167\u5DF2\u6062\u590D\u3002</div>'));
      refresh(ctx2, id, box);
    });
  }
  async function mountSessionCtx(el, ctx2, id) {
    const box = el.querySelector("[data-ctx-panel]");
    if (!box || !id) return;
    if (focusIn(box)) return;
    const cached = PANEL.get(id);
    if (cached && cached.model) box.innerHTML = '<div class="card" data-ctx-card>' + renderSessionCtx(cached.model) + "</div>";
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const r = await ctx2.api.sessionContext(id);
    if ((typeof location !== "undefined" ? location.hash : "") !== stamp) return;
    if (!r.ok) {
      if (!focusIn(box)) {
        box.innerHTML = t('<div class="card"><div class="card-b"><div class="small faint">\u4E0A\u4E0B\u6587\u9762\u6682\u4E0D\u53EF\u8FBE\uFF08') + esc(String(r.error || r.status)) + (r.status === 503 ? t("\u2014\u2014contextidx \u672A\u88C5\u914D\uFF08\u9632\u5FA1\u9762\uFF09") : "") + t("\uFF09\u3002\u6062\u590D\u540E\u91CD\u8FDB\u5373\u73B0\u8BFB\u3002</div></div></div>");
      }
      return;
    }
    if (focusIn(box)) return;
    paint(box, ctx2, id, normalizeContext(r.data));
  }

  // src/screens/session-stages.js
  function normalizeSessionConfig(data) {
    const d = data && typeof data === "object" ? data : {};
    const c = d.config && typeof d.config === "object" ? d.config : {};
    const st = c.stages && typeof c.stages === "object" ? c.stages : {};
    const def2 = {
      s1: "\u7B2C\u4E00\u6BB5\xB7\u80CC\u666F\u88C5\u914D\uFF08LLM \u63CF\u8FF0\u547D\u4EE4 \u2192 tag \u5206\u6790\u5148\u67E5 CtxCache \u2192 \u4E09\u8DEF\u6CE8\u5165 \u2192 builtin RAG \u2192 \u7C7B\u4F3C session \u2192 developer_system \u5B9A\u4F4D\u4ED3\u5E93 \u2192 \u80CC\u666F \u22645k \u9010\u5757\u6EAF\u6E90\uFF09",
      s2: "\u7B2C\u4E8C\u6BB5\xB7\u9884\u66F4\u65B0\u6267\u884C\uFF08\u53CD\u590D ask \u8BE2\u95EE\uFF1B\u4F1A\u8BDD\u6CA1\u6709\u547D\u4EE4\u2014\u2014\u53EA\u4EA7\u6587\u6863\u9884\u66F4\u65B0 draft\uFF1B\u7B49\u5F85\u7528\u6237 apply\uFF09",
      s3: "\u7B2C\u4E09\u6BB5\xB7\u5C5E\u6027\u56DE\u6574\uFF08apply \u89E6\u53D1\uFF1Aref/tag/\u91CD\u8981\u7A0B\u5EA6/metainfo \u9ED8\u8BA4\u81EA\u52A8\u66F4\u65B0\uFF09"
    };
    const scope = (Array.isArray(c.metaScope) ? c.metaScope : []).map(String).filter(Boolean);
    return {
      persistent: d.persistent !== false,
      note: String(d.note || ""),
      stages: {
        s1: String(st.s1 || def2.s1),
        s2: String(st.s2 || def2.s2),
        s3: String(st.s3 || def2.s3)
      },
      backgroundMaxChars: Number(c.backgroundMaxChars) || 5e3,
      blockAttribution: c.blockAttribution !== false,
      metaAutoUpdate: c.metaAutoUpdate !== false,
      metaScope: scope.length ? scope : ["ref", "tag", "importance", "metainfo"],
      provider: String(c.provider || "openai-default"),
      model: String(c.model || "reasoning-tier"),
      updatedAt: String(c.updatedAt || "")
    };
  }
  function normalizeWritebacks(data) {
    const d = data && typeof data === "object" ? data : {};
    return (Array.isArray(d.items) ? d.items : []).map((w) => {
      const row = w && typeof w === "object" ? w : {};
      return {
        at: String(row.at || ""),
        resource: String(row.resource || ""),
        subject: String(row.subject || ""),
        sessionId: String(row.sessionId || ""),
        mode: String(row.mode || "rule-default"),
        importanceDelta: Number(row.importanceDelta) || 0,
        tagsAdded: (Array.isArray(row.tagsAdded) ? row.tagsAdded : []).map(String),
        note: String(row.note || "")
      };
    });
  }
  function deriveStageState(input) {
    const i = input && typeof input === "object" ? input : {};
    const ctxReady = !!i.ctxReady;
    const running = i.running === true;
    const wb = Array.isArray(i.writebacks) && i.writebacks.length ? i.writebacks[0] : null;
    return [
      {
        n: 1,
        name: t("\u7B2C\u4E00\u6BB5\xB7\u80CC\u666F\u88C5\u914D"),
        done: ctxReady,
        active: !ctxReady,
        detail: ctxReady ? t("\u4E0A\u4E0B\u6587\u9762\u5DF2\u73B0\u8BFB\uFF08\u7F13\u5B58 chunks/\u5206\u7247\u89C1\u53F3\u680F\uFF09") : t("\u7B49\u5F85\u4F1A\u8BDD\u4E0A\u4E0B\u6587\u9762\u88C5\u914D")
      },
      {
        n: 2,
        name: t("\u7B2C\u4E8C\u6BB5\xB7\u9884\u66F4\u65B0\u6267\u884C"),
        done: !running && !!wb,
        active: running,
        detail: running ? t("\u4F1A\u8BDD\u8FD0\u884C\u4E2D\uFF08\u53CD\u590D ask \u8BE2\u95EE\uFF1B\u4EA7\u51FA=\u6587\u6863\u9884\u66F4\u65B0 draft\uFF09") : wb ? t("\u9884\u66F4\u65B0\u5DF2 apply\u2014\u2014\u8FDB\u5165\u56DE\u6574") : t("\u7B49\u5F85\u8F93\u5165\u6216\u7B49\u5F85 apply\uFF08\u4F1A\u8BDD\u6CA1\u6709\u547D\u4EE4\u2014\u2014\u53EA\u4EA7 draft\uFF09")
      },
      {
        n: 3,
        name: t("\u7B2C\u4E09\u6BB5\xB7\u5C5E\u6027\u56DE\u6574"),
        done: !!wb,
        active: false,
        detail: wb ? wb.resource + (wb.importanceDelta ? t(" \xB7 importance +") + wb.importanceDelta : "") + (wb.tagsAdded && wb.tagsAdded.length ? t(" \xB7 tag+") + wb.tagsAdded.length : "") : t("apply \u540E\u81EA\u52A8\u89E6\u53D1\uFF08\u9ED8\u8BA4\u56DE\u6574\uFF1Aref/tag/\u91CD\u8981\u7A0B\u5EA6/metainfo\uFF09")
      },
      { wb }
    ];
  }
  function renderStagesCard(cfg, stages, writebacks, persistent) {
    const stageRow = (s) => '<div class="mini-row"><div class="mr-main"><div>' + (s.done ? chip(t("\u5B8C\u6210"), "ok") : s.active ? chip(t("\u8FDB\u884C\u4E2D"), "acc") : chip(t("\u5F85\u89E6\u53D1"), "mute")) + " <b>" + esc(String(s.name || "")) + '</b></div><div class="mr-sub small">' + esc(trunc(String(s.detail || ""), 90)) + "</div></div></div>";
    const wbRows = (writebacks || []).slice(0, 5).map((w) => '<div class="mini-row"><div class="mr-main"><div>' + chip(t("\u81EA\u52A8\u56DE\u6574"), "llm tiny") + ' <span class="mono small">' + esc(trunc(w.resource, 48)) + "</span>" + (w.importanceDelta ? " " + chip("importance +" + w.importanceDelta, "ok tiny") : "") + (w.tagsAdded && w.tagsAdded.length ? w.tagsAdded.map((t2) => chip(t2, "mute tiny")).join(" ") : "") + '</div><div class="mr-sub small">' + esc(w.at || "") + " \xB7 " + esc(w.mode) + (w.note ? " \xB7 " + esc(trunc(w.note, 60)) : "") + "</div></div></div>").join("");
    const disp = (b) => b ? t("\u5F00") : t("\u5173");
    return '<div class="card-h"><span class="h-title">' + esc(t("\u4E09\u6BB5\u5F0F Session")) + '</span><span class="h-right">' + (persistent ? chip(t("\u5DF2\u6301\u4E45\u5316"), "ok tiny") : chip(t("\u7F3A\u7701\uFF08\u672A\u6301\u4E45\u5316\uFF09"), "warn tiny")) + '</span></div><div class="card-b">' + (stages || []).filter((s) => s && s.n).map(stageRow).join("") + '<div class="small faint" style="margin-top:6px">' + esc(t("\u80CC\u666F\u9884\u7B97 ")) + esc(fmtNum(cfg.backgroundMaxChars)) + t(" \u5B57 \xB7 \u9010\u5757\u6EAF\u6E90 ") + esc(disp(cfg.blockAttribution)) + t(" \xB7 \u5C5E\u6027\u56DE\u6574 ") + esc(disp(cfg.metaAutoUpdate)) + t(" \xB7 Provider ") + esc(cfg.provider) + " / " + esc(cfg.model) + "</div>" + (wbRows ? '<h3 class="sec-title" style="margin:10px 0 4px">' + esc(t("\u56DE\u6574\u8D26\uFF08\u7B2C\u4E09\u6BB5\uFF09")) + "</h3>" + wbRows : "") + '<details class="ex-block" style="margin-top:8px"><summary class="ex-sum"><span class="ex-chev">\u25B6</span>' + esc(t("\u4E09\u6BB5\u63A7\u5236\uFF08\u63CF\u8FF0/\u9884\u7B97/\u5F00\u5173\u2014\u2014PUT /v1/session-config\uFF09")) + '</summary><div class="ex-body"><div class="small faint" style="margin:4px 0">' + esc(t("\u4E09\u6BB5\u63CF\u8FF0\u4E0E\u5F00\u5173=\u4E09\u6BB5\u5F0F\u884C\u4E3A\u7684\u627F\u8F7D\u9762\uFF08scheme runtime.context \u540C\u6E90\uFF1B\u6B64\u5904\u4E3A\u6570\u636E\u5E93\u627F\u8F7D\u884C\u2014\u2014\u4FDD\u5B58\u5373\u751F\u6548\uFF09\u3002")) + '</div><textarea class="inp" data-stage-desc="s1" rows="2" style="width:100%">' + esc(cfg.stages.s1) + '</textarea><textarea class="inp" data-stage-desc="s2" rows="2" style="width:100%;margin-top:4px">' + esc(cfg.stages.s2) + '</textarea><textarea class="inp" data-stage-desc="s3" rows="2" style="width:100%;margin-top:4px">' + esc(cfg.stages.s3) + '</textarea><div style="display:grid;grid-template-columns:1fr 1fr;gap:6px;margin-top:6px"><label class="small">' + esc(t("\u80CC\u666F\u9884\u7B97\uFF08\u5B57\uFF09")) + ' <input class="inp" data-stage-num="backgroundMaxChars" type="number" min="500" step="500" value="' + cfg.backgroundMaxChars + '"></label><label class="small">' + esc(t("\u56DE\u6574\u5C5E\u6027\u57DF")) + ' <input class="inp" data-stage-in="metaScope" value="' + esc(cfg.metaScope.join(",")) + '"></label><label class="small">' + esc(t("\u4F1A\u8BDD Provider")) + ' <input class="inp" data-stage-in="provider" value="' + esc(cfg.provider) + '"></label><label class="small">' + esc(t("\u4F1A\u8BDD\u6A21\u578B")) + ' <input class="inp" data-stage-in="model" value="' + esc(cfg.model) + '"></label></div><div style="display:flex;gap:10px;margin-top:6px" class="small"><label><input type="checkbox" data-stage-chk="blockAttribution"' + (cfg.blockAttribution ? " checked" : "") + "> " + esc(t("\u80CC\u666F\u9010\u5757\u6EAF\u6E90")) + '</label><label><input type="checkbox" data-stage-chk="metaAutoUpdate"' + (cfg.metaAutoUpdate ? " checked" : "") + "> " + esc(t("\u7B2C\u4E09\u6BB5\u5C5E\u6027\u81EA\u52A8\u56DE\u6574")) + '</label></div><div style="margin-top:8px;display:flex;gap:8px;align-items:center"><button class="btn btn-secondary sm" data-stage-save>' + esc(t("\u4FDD\u5B58\u4E09\u6BB5\u63A7\u5236")) + '</button><span class="small faint" data-stage-msg></span></div></div></details></div>';
  }
  var STAGES = { model: null, writebacks: [], persistent: false };
  function paint2(box, cfg, stages, writebacks, persistent) {
    box.innerHTML = '<div class="card" data-stages-card>' + renderStagesCard(cfg, stages, writebacks, persistent) + "</div>";
    const save = box.querySelector("[data-stage-save]");
    if (save) save.addEventListener("click", async () => {
      const q = (sel) => box.querySelector(sel);
      const payload = {
        stages: {
          s1: q('[data-stage-desc="s1"]').value,
          s2: q('[data-stage-desc="s2"]').value,
          s3: q('[data-stage-desc="s3"]').value
        },
        backgroundMaxChars: Number(q('[data-stage-num="backgroundMaxChars"]').value) || 5e3,
        blockAttribution: q('[data-stage-chk="blockAttribution"]').checked,
        metaAutoUpdate: q('[data-stage-chk="metaAutoUpdate"]').checked,
        metaScope: q('[data-stage-in="metaScope"]').value.split(/[\s,，、]+/).map((x) => x.trim()).filter(Boolean),
        provider: q('[data-stage-in="provider"]').value.trim() || "openai-default",
        model: q('[data-stage-in="model"]').value.trim() || "reasoning-tier"
      };
      const msg = q("[data-stage-msg]");
      if (msg) msg.textContent = t("\u4FDD\u5B58\u4E2D\u2026");
      const r = await window.__CC_CONSOLE_CTX__.api.saveSessionConfig(payload);
      if (msg) msg.textContent = r.ok ? t("\u5DF2\u4FDD\u5B58\uFF08") + (r.data && r.data.updatedAt || "") + t("\uFF09") : t("\u4FDD\u5B58\u5931\u8D25\uFF1A") + String(r.error || r.status);
      refresh2(box);
    });
  }
  async function refresh2(box) {
    const api2 = typeof window !== "undefined" && window.__CC_CONSOLE_CTX__ && window.__CC_CONSOLE_CTX__.api || null;
    if (!api2 || !api2.sessionConfig) return;
    const [c, w] = await Promise.all([api2.sessionConfig(), api2.sessionWritebacks ? api2.sessionWritebacks(5) : Promise.resolve({ ok: false })]);
    if (!c.ok) return;
    STAGES.model = normalizeSessionConfig(c.data);
    STAGES.persistent = STAGES.model.persistent;
    STAGES.writebacks = w && w.ok ? normalizeWritebacks(w.data) : STAGES.writebacks;
    const ctxCard = document.querySelector("[data-ctx-card]");
    const running = !!(window.__CC_CHAT_STATE__ && window.__CC_CHAT_STATE__.running === true);
    const stages = deriveStageState({ ctxReady: !!ctxCard, running, writebacks: STAGES.writebacks });
    paint2(box, STAGES.model, stages, STAGES.writebacks, STAGES.persistent);
  }
  async function mountSessionStages(el) {
    const box = el.querySelector("[data-stages-panel]");
    if (!box) return;
    await refresh2(box);
  }
  function repaintSessionStages(el) {
    const box = el.querySelector("[data-stages-panel]");
    if (!box || !STAGES.model) return;
    const ctxCard = document.querySelector("[data-ctx-card]");
    const running = !!(window.__CC_CHAT_STATE__ && window.__CC_CHAT_STATE__.running === true);
    const stages = deriveStageState({ ctxReady: !!ctxCard, running, writebacks: STAGES.writebacks });
    paint2(box, STAGES.model, stages, STAGES.writebacks, STAGES.persistent);
  }

  // src/render/refs.js
  var REF_KINDS = "Rule|Goal|Desc|Runbook|DriftReport|DocReview|Agent|ServiceDesc|ServiceInstance|WorkService|ConnectorOps|Provider|Ref|Role|Group|Workflow|Task|Proposal|Notification|DesignDoc|BugDoc|Doc|Template|Event";
  function refRoute(kind, scope, name) {
    const k = encodeURIComponent(kind);
    const s = encodeURIComponent(scope || "global");
    const n = encodeURIComponent(name);
    return "obj/" + k + "/" + s + "/" + n;
  }
  function linkifyRefs(text2, names) {
    const s = String(text2 || "");
    const nameSet = names instanceof Set ? names : Array.isArray(names) ? new Set(names) : null;
    const re = new RegExp("(" + REF_KINDS + ")\\/([A-Za-z0-9_.\\/-]+?)\\/([A-Za-z0-9_.\\-\u4E00-\u9FFF]+)", "g");
    const connRe = /\b(connector|provider):([A-Za-z0-9_.-]+)/g;
    let html3 = "";
    let last = 0;
    const spans = [];
    let m;
    while ((m = re.exec(s)) !== null) {
      spans.push({ i: m.index, len: m[0].length, kind: m[1], scope: m[2], name: m[3] });
    }
    while ((m = connRe.exec(s)) !== null) {
      spans.push({ i: m.index, len: m[0].length, conn: m[1], cname: m[2] });
    }
    spans.sort((a, b) => a.i - b.i);
    for (const sp of spans) {
      if (sp.i < last) continue;
      html3 += esc(s.slice(last, sp.i));
      if (sp.conn) {
        html3 += '<span class="ref-chip" data-go="connectors">' + esc(sp.conn + ":" + sp.cname) + "</span>";
      } else {
        const route = refRoute(sp.kind, sp.scope, sp.name);
        html3 += '<span class="ref-chip" data-go="' + esc(route) + '" title="' + esc(t("\u67E5\u770B\u5BF9\u8C61")) + '">' + esc(sp.kind + "/" + sp.scope + "/" + sp.name) + "</span>";
      }
      last = sp.i + sp.len;
    }
    html3 += esc(s.slice(last));
    if (nameSet && nameSet.size) {
      for (const nm of nameSet) {
        const needle = esc(nm);
        if (!needle || !html3.includes(needle)) continue;
        html3 = html3.split(needle).join(
          '<span class="ref-chip" data-go="' + esc(refRoute("Doc", "global", nm)) + '" data-ref-name="' + esc(nm) + '">' + needle + "</span>"
        );
      }
    }
    return html3;
  }

  // src/screens/chat-views.js
  var N_ICONS = {
    job: '<rect x="2" y="3.5" width="12" height="2.2" rx="1.1"/><rect x="2" y="7.4" width="8.5" height="2.2" rx="1.1"/><rect x="2" y="11.3" width="5.5" height="2.2" rx="1.1"/>',
    plugin: '<rect x="3" y="3" width="10" height="10" rx="2.5"/><path d="M8 6.2v3.6M6.2 8h3.6"/>',
    schedule: '<circle cx="8" cy="8" r="5.5"/><path d="M8 5.2V8l2 1.6"/>',
    subagent: '<circle cx="5.6" cy="5.8" r="2.1"/><path d="M2.4 13c.5-2.4 1.8-3.4 3.2-3.4s2.7 1 3.2 3.4"/><circle cx="11" cy="6.6" r="1.7"/><path d="M10.4 9.7c1.6.4 2.5 1.5 2.9 3.3"/>',
    goal: '<circle cx="8" cy="8" r="5.5"/><circle cx="8" cy="8" r="2.2"/>',
    agent: '<path d="M2.5 8 13.5 3 11 13 8 9.5 2.5 8z"/><path d="M8 9.5 13.5 3"/>',
    team: '<circle cx="5.6" cy="5.8" r="2.1"/><path d="M2.4 13c.5-2.4 1.8-3.4 3.2-3.4s2.7 1 3.2 3.4"/><circle cx="11" cy="6.6" r="1.7"/><path d="M10.4 9.7c1.6.4 2.5 1.5 2.9 3.3"/>',
    github: '<circle cx="4.5" cy="4" r="1.7"/><circle cx="4.5" cy="12" r="1.7"/><circle cx="11.5" cy="8" r="1.7"/><path d="M4.5 5.7v4.6M6.2 8h3.6"/>',
    webhook: '<circle cx="8" cy="8" r="5.5"/><path d="M2.5 8h11M8 2.5c2.6 2.6 2.6 8.4 0 11-2.6-2.6-2.6-8.4 0-11"/>',
    retry: '<path d="M3 8a5 5 0 1 1 1.5 3.5"/><path d="M3 12.5V8.5h4"/>',
    inject: '<rect x="2.5" y="5.5" width="11" height="8" rx="2"/><path d="M8 2v5M6 5l2 2 2-2"/>',
    recall: '<path d="M5 2.5h6a1 1 0 0 1 1 1V14l-4-2.6L4 14V3.5a1 1 0 0 1 1-1z"/>'
  };
  function noticeIcon(id) {
    return '<svg class="chat-n-ico" viewBox="0 0 16 16" aria-hidden="true">' + (N_ICONS[id] || N_ICONS.inject) + "</svg>";
  }
  function clockAttr(n) {
    const c = formatClock(n.time);
    return c ? ' title="' + esc(c) + '"' : "";
  }
  function noticeHtml(n, names) {
    const head = "<summary>" + noticeIcon(n.trigger && n.trigger.id) + '<span class="chat-n-title">' + esc(n.title) + "</span>" + (n.producer ? '<span class="chat-n-sep">\xB7</span><span class="chat-n-producer mono">' + esc(n.producer) + "</span>" : "") + (n.summary ? '<span class="chat-n-sep">\xB7</span><span class="chat-n-summary">' + esc(n.summary) + "</span>" : "") + (n.time ? '<time class="chat-n-time">' + esc(formatClock(n.time)) + "</time>" : "") + '<span class="ex-chev">\u25B8</span></summary>';
    const body = '<div class="ex-body">' + (n.waking ? '<div class="chat-n-exp">' + esc(t("\u8FD9\u6761\u901A\u77E5\u89E6\u53D1\u4E86\u672C\u8F6E\u56DE\u590D\u3002")) + "</div>" : "") + '<div class="chat-n-text">' + (n.text ? linkifyRefs(n.text, names) : '<span class="faint">\u2026</span>') + "</div></div>";
    return '<details class="chat-notice' + (n.waking ? " chat-trigger" : "") + '" data-node="' + esc(nodeKey(n)) + '"' + clockAttr(n) + ">" + head + body + "</details>";
  }
  function reasoningFold(reasoning, names) {
    const preview = String(reasoning || "").split("\n").map((l) => l.replace(/\*\*/g, "")).find((l) => l.trim()) || "";
    return '<details class="chat-fold"><summary><span class="ex-chev">\u25B8</span>' + esc(t("\u601D\u8003")) + (preview ? '<span class="chat-think-preview">' + esc(trunc(preview.trim(), 72)) + "</span>" : "") + '</summary><div class="ex-body">' + linkifyRefs(reasoning, names) + "</div></details>";
  }
  function usagePill(usage) {
    const total = usageTotal(usage);
    if (!total) return "";
    const detail = usageDetail(usage);
    return '<span class="chat-usage" title="' + esc(detail || t("\u672C\u8F6E\u7528\u91CF")) + '">' + esc(t("\u7528\u91CF {total}").replace("{total}", fmtNum(total))) + "</span>";
  }
  function toolHtml(n) {
    const st = n.state || "ok";
    let stateHtml = "";
    if (st === "running") stateHtml = '<span class="chat-tool-run">' + esc(t("\u8FD0\u884C\u4E2D")) + "</span>";
    else if (st === "error") {
      stateHtml = '<span class="chat-tool-err">' + esc(t("\u5931\u8D25")) + "</span>";
      const code = Number(n.meta && n.meta.exitCode);
      if (n.meta && Number.isFinite(code) && code !== 0) {
        stateHtml += '<span class="chat-tool-exit">' + esc(t("\u9000\u51FA\u7801 {code}").replace("{code}", String(n.meta.exitCode))) + "</span>";
      } else if (n.meta && typeof n.meta.signal === "string" && n.meta.signal) {
        stateHtml += '<span class="chat-tool-exit">' + esc(t("\u4FE1\u53F7 {signal}").replace("{signal}", n.meta.signal)) + "</span>";
      }
    } else if (st === "stopped") stateHtml = '<span class="chat-tool-stop">' + esc(t("\u5DF2\u505C\u6B62")) + "</span>";
    const head = '<summary><span class="ex-chev">\u25B8</span>' + (n.title ? '<span class="chat-tool-title">' + esc(n.title) + "</span>" : "<span>" + esc(t("\u5DF2\u8C03\u7528")) + '</span> <span class="chat-tool-name">' + esc(n.name) + "</span>") + (n.summary ? '<span class="chat-tool-arg">' + esc(n.summary) + "</span>" : "") + (n.metaLine ? '<span class="chat-tool-metaline">' + esc(n.metaLine) + "</span>" : "") + " " + stateHtml + "</summary>";
    const linesPrev = n.meta && Array.isArray(n.meta.lines) && n.meta.lines.length ? '<div class="chat-tool-lines">' + n.meta.lines.slice(0, 24).map((l) => '<span class="chat-tool-ln">' + esc(String((l && l.number) != null ? l.number : "")) + '</span><span class="chat-tool-lt">' + esc(String((l && l.text) != null ? l.text : "")) + "</span>").join("") + "</div>" : "";
    const errLine = n.error && (n.error.name || n.error.code) ? "\n" + t("\u5931\u8D25\u539F\u56E0\uFF1A") + [n.error.name, n.error.code].filter(Boolean).join("/") : "";
    const detail = linesPrev + '<div class="chat-tool-io">' + esc(trunc(String(n.args || ""), 800)) + (n.result ? "\n" + t("\u2500\u2500\u2500\u2500 \u7ED3\u679C \u2500\u2500\u2500\u2500") + "\n" + esc(trunc(String(n.result), 1200)) : "") + errLine + "</div>";
    return '<details class="chat-fold chat-tool" data-node="' + esc(nodeKey(n)) + '" data-state="' + esc(st) + '"' + clockAttr(n) + ">" + head + '<div class="ex-body">' + detail + "</div></details>";
  }
  function turnTailHtml(n) {
    const tt = turnTailOf(n.reason, n.durationMs);
    return '<div class="chat-turntail-wrap" data-node="' + esc(nodeKey(n)) + '"' + clockAttr(n) + '><div class="chat-turntail ' + esc(tt.cls) + '"><span class="chat-tt-dot"></span><span class="chat-tt-word">' + esc(tt.word) + "</span>" + (n.time ? '<time class="chat-n-time">' + esc(formatClock(n.time)) + "</time>" : "") + "</div>" + (tt.hint ? '<div class="chat-tt-hint">' + esc(tt.hint) + "</div>" : "") + "</div>";
  }
  function todoHtml(n) {
    const m = todoModel(n.todos);
    if (!m.total) return "";
    const dot = (s) => '<span class="chat-todo-dot ' + (s === "completed" ? "done" : s === "in_progress" ? "act" : "idle") + '"></span>';
    return '<details class="chat-fold chat-todo" data-node="' + esc(nodeKey(n)) + '"' + clockAttr(n) + '><summary><span class="ex-chev">\u25B8</span>' + esc(t("\u66F4\u65B0\u4EFB\u52A1\u6E05\u5355")) + ' <span class="chat-todo-count">' + esc(t("{done}/{total} \u5DF2\u5B8C\u6210").replace("{done}", String(m.done)).replace("{total}", String(m.total))) + "</span>" + (m.active ? '<span class="chat-todo-active">' + esc(trunc(m.active, 44)) + "</span>" : "") + '</summary><div class="ex-body chat-todo-list">' + m.rows.map((r) => '<div class="chat-todo-row">' + dot(r.status) + "<span>" + esc(r.content) + "</span></div>").join("") + "</div></details>";
  }
  function retryHtml(n) {
    const d = n.data && typeof n.data === "object" ? n.data : {};
    const fail = d.failure && typeof d.failure === "object" ? String(d.failure.message || "") : String(d.message || "");
    const label = typeof d.label === "string" && d.label ? d.label : t("\u6B63\u5728\u91CD\u8BD5\u6A21\u578B\u8BF7\u6C42");
    const extra = [];
    if (Number.isFinite(Number(d.attempt)) && Number.isFinite(Number(d.maximum))) {
      extra.push(t("{label}\uFF08{retry}/{maximum}\uFF09").replace("{label}", label).replace("{retry}", String(d.attempt)).replace("{maximum}", String(d.maximum)));
      if (Number.isFinite(Number(d.seconds))) extra.push(String(d.seconds) + "s");
    } else extra.push(label);
    return '<div class="chat-sysrow chat-retry" data-node="' + esc(nodeKey(n)) + '"' + clockAttr(n) + ">" + esc(extra.join(" \xB7 ")) + (fail ? " \xB7 " + esc(t("\u5931\u8D25\u539F\u56E0\uFF1A") + trunc(fail, 120)) : "") + "</div>";
  }
  function nodeHtml(n, names) {
    if (n.kind === "user") {
      const pending = n.pending ? '<div class="chat-pending">' + esc(t("\u9001\u8FBE\u4E2D\u2026")) + "</div>" : n.injected ? '<div class="chat-pending">' + esc(t("\u6CE8\u5165\u4E0A\u4E0B\u6587")) + "</div>" : "";
      const err = n.error ? '<div class="chat-err"><span>' + esc(t("\u9001\u8FBE\u5931\u8D25\uFF08") + String(n.error) + t("\uFF09")) + '</span><button class="btn btn-ghost sm" data-cc-retry="' + esc(String(n.requestId || "")) + '">' + esc(t("\u91CD\u8BD5")) + "</button></div>" : "";
      return '<div class="chat-msg-user" data-node="' + esc(nodeKey(n)) + '"' + clockAttr(n) + '><div class="chat-bubble">' + (n.text ? renderMarkdown(n.text) : '<span class="faint">\u2026</span>') + "</div>" + (pending || err ? "<div>" + pending + err + "</div>" : "") + "</div>";
    }
    if (n.kind === "notice") return noticeHtml(n, names);
    if (n.kind === "assistant") {
      if (!n.text && !n.reasoning) return "";
      const reasoning = n.reasoning ? reasoningFold(n.reasoning, names) : "";
      const body = n.text ? '<div class="chat-md">' + renderMarkdown(n.text) + "</div>" : "";
      return '<div class="chat-msg-ai" data-node="' + esc(nodeKey(n)) + '"' + clockAttr(n) + ">" + reasoning + body + usagePill(n.usage) + "</div>";
    }
    if (n.kind === "tool") return toolHtml(n);
    if (n.kind === "turntail") return turnTailHtml(n);
    if (n.kind === "todo") return todoHtml(n);
    if (n.kind === "retry") return retryHtml(n);
    const label = (n.dev ? t("\u5F00\u53D1\u8005") + " \xB7 " : "") + trunc(String(n.text || ""), 200);
    return '<div class="chat-sysrow" data-node="' + esc(nodeKey(n)) + '"' + clockAttr(n) + ">" + esc(label) + "</div>";
  }

  // src/screens/chat-refs.js
  function looksLikePath(s) {
    const t2 = String(s || "").trim().replace(/[),.;:」』）】]+$/, "");
    if (!t2 || t2.length > 160 || /\s/.test(t2) || t2.includes("`") || t2.includes("*")) return false;
    if (/[“”'"<>|\\]/.test(t2)) return false;
    if (!/[A-Za-z]/.test(t2)) return false;
    if (/^(?:~\/|\.{1,2}\/|(?:[\w@.-]+\/)+)[\w@.-]+$/.test(t2)) {
      const segs = t2.split("/");
      const last = segs[segs.length - 1];
      if (!/\.[A-Za-z][A-Za-z0-9]{0,7}$/.test(last) && !/^(?:~\/|\.{1,2}\/)/.test(t2)) return false;
      return true;
    }
    return /\.(md|markdown|yaml|yml|json|ts|tsx|js|mjs|go|py|sh|toml|css|html|sql)$/i.test(t2);
  }
  function extractFilePaths(text2, cap = 8) {
    const s = String(text2 || "");
    const out = [];
    const seen = /* @__PURE__ */ new Set();
    const re = /(?:~\/|\.{1,2}\/)?(?:[\w@.-]+\/)+[\w@.-]+|[\w@.-]+\.(?:md|markdown|yaml|yml|json|ts|tsx|js|mjs|go|py|sh|toml|css|html|sql)/g;
    let m;
    while ((m = re.exec(s)) !== null && out.length < cap * 3) {
      const p = m[0].replace(/[),.;:」』）】]+$/, "");
      if (!looksLikePath(p) || seen.has(p)) continue;
      seen.add(p);
      out.push(p);
      if (out.length >= cap) break;
    }
    return out;
  }
  function extractDocRefs(text2, cap = 12) {
    const s = String(text2 || "");
    const re = new RegExp("(" + REF_KINDS + ")\\/([A-Za-z0-9_.\\/-]+?)\\/([A-Za-z0-9_.\\-\u4E00-\u9FFF]+)", "g");
    const out = [];
    const seen = /* @__PURE__ */ new Set();
    let m;
    while ((m = re.exec(s)) !== null && out.length < cap) {
      const key = m[0];
      if (seen.has(key)) continue;
      seen.add(key);
      out.push({ type: "obj", kind: m[1], scope: m[2], name: m[3], label: key });
    }
    return out;
  }
  var ARGS_PATH_KEYS = ["file_path", "path", "note_path", "cwd"];
  function extractRefs(nodes) {
    const docs = [];
    const docSeen = /* @__PURE__ */ new Set();
    const files = [];
    const fileSeen = /* @__PURE__ */ new Set();
    const addDoc = (d) => {
      const key = d.kind + "/" + d.scope + "/" + d.name;
      if (docSeen.has(key) || docs.length >= 12) return;
      docSeen.add(key);
      docs.push({ ...d, label: key });
    };
    const addFile = (p) => {
      if (!p || fileSeen.has(p) || files.length >= 8) return;
      fileSeen.add(p);
      files.push({ type: "file", path: p, label: p });
    };
    for (const n of Array.isArray(nodes) ? nodes : []) {
      if (!n || typeof n !== "object") continue;
      if (typeof n.text === "string" && n.text) {
        extractDocRefs(n.text).forEach(addDoc);
        extractFilePaths(n.text).forEach(addFile);
      }
      if (n.kind === "tool" && n.args) {
        try {
          const a = JSON.parse(String(n.args));
          if (a && typeof a === "object") {
            for (const k of ARGS_PATH_KEYS) {
              if (typeof a[k] === "string" && looksLikePath(a[k])) {
                addFile(a[k].trim());
                break;
              }
            }
          }
        } catch {
        }
      }
    }
    return { docs, files };
  }
  var refChipHtml = (r, i, group) => '<button class="chat-ref" data-cc-ref-g="' + esc(group) + '" data-cc-ref-i="' + i + '" title="' + esc(r.label) + '"><span class="chat-ref-name">' + esc(trunc(r.label, 56)) + "</span>" + (r.score ? '<span class="chat-ref-score">' + esc(String(Math.round(r.score))) + "</span>" : "") + '<span class="chat-ref-open">' + esc(t("\u6253\u5F00")) + "</span></button>";
  function refsStripHtml(refs) {
    const r = refs && typeof refs === "object" ? refs : {};
    const cited = [].concat(Array.isArray(r.docs) ? r.docs : [], Array.isArray(r.files) ? r.files : []);
    const chunks = Array.isArray(r.chunks) ? r.chunks : [];
    if (!cited.length && !chunks.length) return "";
    const groups = [];
    if (cited.length) {
      groups.push('<span class="chat-refs-h">' + esc(t("\u5F15\u7528")) + "</span>" + cited.map((x, i) => refChipHtml(x, i, "cited")).join(""));
    }
    if (chunks.length) {
      groups.push('<span class="chat-refs-h">' + esc(t("\u4E0A\u4E0B\u6587")) + "</span>" + chunks.map((x, i) => refChipHtml(x, i, "chunk")).join(""));
    }
    return '<div class="chat-refs">' + groups.join("") + "</div>";
  }
  function refTarget(refs, g, i) {
    const r = refs && typeof refs === "object" ? refs : {};
    const gi = Number(i);
    if (!Number.isFinite(gi)) return null;
    if (g === "cited") {
      const cited = [].concat(Array.isArray(r.docs) ? r.docs : [], Array.isArray(r.files) ? r.files : []);
      return cited[gi] || null;
    }
    if (g === "chunk") {
      return (Array.isArray(r.chunks) ? r.chunks : [])[gi] || null;
    }
    return null;
  }
  function modalShell() {
    return '<div class="chat-modal" data-cc-modal hidden><div class="chat-modal-mask" data-cc-modal-close></div><div class="chat-modal-panel" role="dialog" aria-modal="true"><div class="chat-modal-head"><span class="chat-modal-title mono" data-cc-modal-title></span><span class="chat-modal-acts"><button class="btn btn-ghost sm" data-cc-modal-go hidden>' + esc(t("\u5728\u6587\u6863\u5E93\u4E2D\u6253\u5F00")) + '</button><button class="btn btn-ghost sm" data-cc-modal-close>' + esc(t("\u5173\u95ED")) + '</button></span></div><div class="chat-modal-body" data-cc-modal-body></div></div></div>';
  }
  async function fetchObjContent(ctxc, ref) {
    const r = await ctxc.api.object(ref.kind, ref.scope, ref.name);
    if (!r || !r.ok && r.status !== 200) return { ok: false, note: t("\u5BF9\u8C61\u9762\u6682\u4E0D\u53EF\u8BFB\uFF08") + String(r && (r.error || r.status) || t("\u7F51\u7EDC")) + t("\uFF09") };
    const d = r.data && typeof r.data === "object" ? r.data : {};
    const content = d.content != null ? String(d.content) : "";
    if (!content.trim()) return { ok: false, note: t("\u672C\u6587\u6863\u6682\u65E0\u6B63\u6587\u3002") };
    return { ok: true, content };
  }
  async function fetchFileContent(ctxc, ref, tree) {
    const p = String(ref.path || "");
    const rows = tree && Array.isArray(tree.docs) ? tree.docs : [];
    const hit = rows.find((d2) => d2 && d2.path === p) || rows.find((d2) => d2 && (d2.path === p || d2.path.endsWith("/" + p) || p.endsWith("/" + String(d2.path))));
    if (!hit || !hit.repo) return { ok: false, note: t("\u672A\u5728\u6587\u6863\u5E93\u547D\u4E2D\uFF08design-docs \u767D\u540D\u5355\u4ED3\u6811\u65E0\u6B64\u8DEF\u5F84\uFF09\u2014\u2014\u6765\u6E90\u6458\u5F55\u89C1\u5BF9\u8BDD\u539F\u6587\u3002") };
    const r = await ctxc.api.designDoc(hit.repo, hit.path);
    if (!r || !r.ok && r.status !== 200) return { ok: false, note: t("\u6587\u6863\u76F4\u8BFB\u5931\u8D25\uFF08") + String(r && (r.error || r.status) || t("\u7F51\u7EDC")) + t("\uFF09") };
    const d = r.data && typeof r.data === "object" ? r.data : {};
    const content = d.content != null ? String(d.content) : typeof r.data === "string" ? r.data : "";
    if (!content.trim()) return { ok: false, note: t("\u672C\u6587\u6863\u6682\u65E0\u6B63\u6587\u3002") };
    return { ok: true, content, repo: hit.repo, path: hit.path };
  }
  function renderDocBody(content, metaRef) {
    const meta = metaRef && metaRef.kind || metaRef && metaRef.scope != null || metaRef && metaRef.name ? frontMetaOf(content, metaRef && metaRef.kind, metaRef && metaRef.scope, metaRef && metaRef.name) : frontMetaOf(content);
    return metaHeaderHtml(meta, {}) + '<div data-md-render="1">' + renderMarkdown(splitFrontmatter(content).body) + "</div>" + rawDisclosure(t("\u539F\u6587"), content);
  }
  function mountRefs(el, ctxc) {
    const slot = el.querySelector("[data-cc-refs]");
    if (!slot || !ctxc) return () => {
    };
    if (!el.querySelector("[data-cc-modal]")) slot.insertAdjacentHTML("afterend", modalShell());
    const modal = el.querySelector("[data-cc-modal]");
    const body = modal.querySelector("[data-cc-modal-body]");
    const title = modal.querySelector("[data-cc-modal-title]");
    const goBtn = modal.querySelector("[data-cc-modal-go]");
    let tree = null;
    let treeTried = false;
    const closeModal = () => {
      modal.hidden = true;
      body.innerHTML = "";
    };
    modal.querySelector("[data-cc-modal-close].btn").addEventListener("click", closeModal);
    modal.querySelector(".chat-modal-mask").addEventListener("click", closeModal);
    el.addEventListener("keydown", (ev) => {
      if (ev.key === "Escape" && !modal.hidden) closeModal();
    });
    const openTarget = async (ref) => {
      if (!ref) return;
      modal.hidden = false;
      title.textContent = ref.label || (ref.type === "file" ? ref.path : "");
      goBtn.hidden = true;
      body.innerHTML = '<div class="small faint">' + esc(t("\u73B0\u8BFB\u4E2D\u2026")) + "</div>";
      try {
        const res = ref.type === "file" ? await fetchFileContent(ctxc, ref, tree) : await fetchObjContent(ctxc, ref);
        if (modal.hidden) return;
        if (res.ok) {
          body.innerHTML = renderDocBody(res.content, ref.type === "obj" ? { kind: ref.kind, scope: ref.scope, name: ref.name } : null);
          if (res.repo && res.path) goBtn.dataset.ccGo = "designdoc/" + encodeURIComponent(res.repo) + "/" + encodeURIComponent(res.path);
          else if (ref.type === "obj") goBtn.dataset.ccGo = refRoute(ref.kind, ref.scope, ref.name);
          else goBtn.dataset.ccGo = "";
          goBtn.hidden = !goBtn.dataset.ccGo;
        } else {
          body.innerHTML = '<div class="chat-modal-miss"><div class="mono small">' + esc(ref.label || "") + '</div><div class="small faint">' + esc(res.note) + "</div></div>";
        }
      } catch (e) {
        body.innerHTML = '<div class="small faint">' + esc(t("\u73B0\u8BFB\u5931\u8D25\uFF08") + String(e && e.message || e) + t("\uFF09")) + "</div>";
      }
    };
    el.addEventListener("click", (ev) => {
      const chip2 = ev.target.closest("[data-cc-ref-g]");
      if (chip2 && slot.contains(chip2)) {
        openTarget(refTarget(el.__chatRefs || {}, chip2.getAttribute("data-cc-ref-g"), chip2.getAttribute("data-cc-ref-i")));
        return;
      }
      const mention = ev.target.closest("[data-cc-mention]");
      if (mention) {
        openTarget({ type: "file", path: String(mention.getAttribute("data-cc-mention") || ""), label: String(mention.getAttribute("data-cc-mention") || "") });
        return;
      }
      const go = ev.target.closest("[data-cc-modal-go]");
      if (go && go.dataset.ccGo && ctxc.go) ctxc.go(go.dataset.ccGo);
    });
    const paint3 = (refs) => {
      el.__chatRefs = refs || {};
      if (slot) slot.innerHTML = refsStripHtml(refs);
      if (!treeTried && ctxc.api && typeof ctxc.api.designDocs === "function") {
        treeTried = true;
        ctxc.api.designDocs().then((r) => {
          if (r && (r.ok || r.status === 200)) tree = r.data || r;
        }).catch(() => {
        });
      }
    };
    return paint3;
  }
  function classifyMentions(root) {
    if (!root || typeof root.querySelectorAll !== "function") return root;
    root.querySelectorAll(".chat-md code:not(.md-mention)").forEach((c) => {
      const t2 = String(c.textContent || "").trim();
      if (t2 && looksLikePath(t2) && t2.length <= 160) {
        c.classList.add("md-mention");
        c.setAttribute("data-cc-mention", t2);
      }
    });
    return root;
  }

  // src/screens/chat.js
  function faces3() {
    return typeof window !== "undefined" && window.__CC && window.__CC.chatFaces || null;
  }
  function nextBatch(nodes, paintedKeys, firstPaint) {
    const rows = Array.isArray(nodes) ? nodes : [];
    if (firstPaint) return rows;
    const out = rows.filter((n) => !paintedKeys.has(nodeKey(n)));
    return out;
  }
  function modeForState(st) {
    return !!(st && st.running === true) ? "steer" : "queue";
  }
  function dayGroupOf(updatedAt, nowMs) {
    const t2 = Number(updatedAt) || 0;
    if (!t2) return "earlier";
    const now = Number(nowMs) || Date.now();
    const startOf = new Date(now);
    startOf.setHours(0, 0, 0, 0);
    return now - t2 < 24 * 3600 * 1e3 && startOf.getTime() <= t2 ? "today" : "earlier";
  }
  var CHAT = {
    sessionId: "",
    // 当前会话（宿主发 id——零手填）
    timer: 0,
    // 轮询定时器（导航离开即清）
    pending: [],
    // 待达回声 {text, requestId, at}（C-05 按 requestId 销账）
    lastSeq: 0,
    // 已渲染转录尾（增量分派锚）
    painted: /* @__PURE__ */ new Set(),
    // 已渲染节点键（nodeKey 族）
    seenSeq: /* @__PURE__ */ new Map(),
    // 会话 id → 已见尾 seq（未读点 highwater）
    state: null,
    // 最近一次 chat-state 词形
    el: null,
    // 当前屏根（路由切换换引用）
    lastNodes: [],
    // 最近映射节点（B1 引用抽取源）
    refs: {},
    // 引用条模型 {docs, files, chunks}（B2）
    refsChunks: [],
    // ContextPack chunks（B4 同源，会话切换时懒拉）
    paintRefs: null
    // mountRefs 返回的条重绘 fn（B3 挂载后可用）
  };
  var LAST = { route: "", paramKey: "", locale: "" };
  function pendingNodes() {
    return CHAT.pending.map((p) => ({ kind: "user", seq: Number.MAX_SAFE_INTEGER, text: p.text, pending: true, requestId: p.requestId }));
  }
  function liveRowHtml(st) {
    if (!st || st.running !== true) return "";
    const tool = String(st.currentTool || "");
    let dur = "";
    if (tool && Number(st.currentToolSince) > 0) {
      const s = Math.max(0, Math.round((Date.now() - Number(st.currentToolSince)) / 1e3));
      dur = "\uFF08" + esc(t("\u5DF2 {n} \u79D2").replace("{n}", String(s))) + "\uFF09";
    }
    const qd = Number(st.queueDepth) || 0;
    const sd = Number(st.steerDepth) || 0;
    const queue = qd ? ' <span class="chip acc tiny">' + esc(t("\u6392\u961F {n}").replace("{n}", String(qd))) + "</span>" : "";
    const steer = sd ? ' <span class="chip llm tiny">' + esc(t("\u8FFD\u52A0 {n}").replace("{n}", String(sd))) + "</span>" : "";
    const body = tool ? t("\u6B63\u5728\u8C03\u7528") + ' <span class="chat-live-tool">' + esc(tool) + "</span>\u2026 " + dur : t("\u6DF1\u5EA6\u6C42\u7D22\u4E2D\u2026");
    return '<div class="chat-live"><span class="lamp pulse-ok"></span><span>' + body + "</span>" + queue + steer + "</div>";
  }
  function emptyStateHtml() {
    const samples = [
      t("\u8FDB\u884C\u4EE3\u7801\u5BA1\u6838\uFF0C\u770B\u770B\u5B9E\u9645\u548C\u6587\u6863\u5DEE\u8DDD\u6709\u591A\u5927\uFF0C\u5E76\u8FDB\u884C\u4FEE\u590D"),
      t("\u627E\u51FA\u6240\u6709 drift \u7684\u6587\u6863\u5E76\u751F\u6210\u5BF9\u8D26\u6E05\u5355"),
      t("\u628A\u5F53\u524D\u5FAA\u73AF\u7684\u72B6\u6001\u5199\u6210\u4E00\u4EFD SessionLog \u6458\u8981"),
      t("\u68C0\u67E5 workflow \u5F15\u64CE\u7684\u8FD0\u884C\u8BB0\u5F55\uFF0C\u6807\u51FA\u5931\u8D25\u539F\u56E0")
    ];
    return '<div class="chat-empty"><div class="chat-empty-hi">' + esc(t("\u6709\u4EC0\u4E48\u53EF\u4EE5\u5E2E\u5FD9\u7684\uFF1F")) + '</div><div class="chat-empty-sub">' + esc(t("\u627E\u6587\u6863\uFF0F\u66F4\u65B0\u6587\u6863\uFF0F\u63D0\u4EA4\u9700\u6C42\u2014\u2014\u4F1A\u8BDD\u7531\u5BBF\u4E3B\u5EFA\u6863\uFF0C\u8F6C\u5F55\u5373\u771F\u5B9E journal\u3002")) + '</div><div class="chat-samples">' + samples.map((s) => '<button class="chat-sample" data-cc-sample>' + esc(s) + "</button>").join("") + "</div></div>";
  }
  function stateChipsHtml(st) {
    if (!st) return "";
    const running = st.running === true;
    const chips = [running ? chip(t("\u8FD0\u884C\u4E2D"), "acc") : chip(t("\u7A7A\u95F2"), "mute")];
    const qd = Number(st.queueDepth) || 0;
    if (qd) chips.push(chip(t("\u6392\u961F {n}").replace("{n}", String(qd)), "llm"));
    return chips.join(" ");
  }
  function sessionRowHtml2(s, cur, highwater) {
    const id = String(s.sessionId || "");
    const seen = Number(highwater instanceof Map ? highwater.get(id) : 0) || 0;
    const unread = !s.blank && Number(s.updatedAt) > seen && id !== cur;
    const live = s.running === true;
    const dot = live ? " live" : unread ? " unread" : "";
    const title = String(s.title || "");
    return '<div class="chat-sess' + (id === cur ? " active" : "") + '" data-sess-pick="' + esc(id) + '" role="link"><span class="chat-dot' + dot + '"></span><div class="chat-sess-main"><div class="chat-sess-id">' + esc(trunc(title || id, 26)) + "</div>" + (title ? '<div class="chat-sess-sub mono">' + esc(trunc(id, 26)) + "</div>" : "") + '</div><div class="chat-sess-side">' + (live ? chip(t("\u8FD0\u884C\u4E2D"), "ok") : s.blank ? chip(t("\u7A7A\u767D"), "mute") : "") + (s.updatedAt ? '<span class="chat-sess-time">' + esc(relAge(Number(s.updatedAt)) || "") + "</span>" : "") + "</div></div>";
  }
  function sessionsHtml(items, cur, seenSeq) {
    const rows = (Array.isArray(items) ? items : []).filter((s) => s && s.sessionId);
    if (!rows.length) return '<div class="chat-sess-day">' + esc(t("\u4F1A\u8BDD\u6E05\u5355\u4E3A\u7A7A\u2014\u2014\u70B9\u300C\u65B0\u4F1A\u8BDD\u300D\u7531\u5BBF\u4E3B\u5EFA\u6863\u3002")) + "</div>";
    const today = rows.filter((s) => dayGroupOf(s.updatedAt) === "today");
    const earlier = rows.filter((s) => dayGroupOf(s.updatedAt) !== "today");
    const out = [];
    if (today.length) {
      out.push('<div class="chat-sess-day">' + esc(t("\u4ECA\u5929")) + "</div>");
      out.push(today.map((s) => sessionRowHtml2(s, cur, seenSeq)).join(""));
    }
    if (earlier.length) {
      out.push('<div class="chat-sess-day">' + esc(t("\u66F4\u65E9")) + "</div>");
      out.push(earlier.map((s) => sessionRowHtml2(s, cur, seenSeq)).join(""));
    }
    return out.join("");
  }
  function composerHtmlInner(can, running) {
    const mode = running ? "steer" : "queue";
    return '<div class="chat-comp"><textarea class="chat-text" data-cc-text rows="1" placeholder="' + esc(t("\u63D0\u4EA4\u9700\u6C42\uFF0F\u627E\u6587\u6863\uFF0F\u66F4\u65B0\u6587\u6863\u2026\u2026")) + '"></textarea><button class="chat-send" data-cc-send data-mode="' + mode + '"' + (can ? "" : " disabled") + ' title="' + esc(running ? t("\u505C\u6B62") : t("\u53D1\u9001")) + '"><span class="chat-send-ico">' + (running ? "\u25A0" : "\u2191") + '</span></button></div><div class="cc-comp-tools"><label class="btn btn-ghost sm cc-img-btn" title="' + esc(t("\u9644\u52A0\u56FE\u7247\uFF08\u22648MB\uFF09")) + '">\u{1F5BC} ' + esc(t("\u56FE\u7247")) + '<input type="file" accept="image/png,image/jpeg,image/gif,image/webp,image/svg+xml" data-cc-img hidden></label><span class="cc-img-previews" data-cc-img-previews></span></div><div class="chat-foot">' + esc(t("Enter \u53D1\u9001 \xB7 Shift+Enter \u6362\u884C")) + "</div>";
  }
  function screenShell(can) {
    return '<div class="chat-screen"><div class="chat-top"><div class="chat-top-title"><h1 class="chat-h1">' + esc(t("\u4F1A\u8BDD")) + '</h1><div class="dh-sub">' + esc(t("\u81EA\u6709\u5BF9\u8BDD\u9762\u2014\u2014\u771F\u5B9E\u8F6C\u5F55\u4E0E\u53D1\u9001\uFF1B\u4F1A\u8BDD id \u7531\u5BBF\u4E3B\u53D1\uFF0C\u6C38\u4E0D\u624B\u586B")) + '</div></div><div class="chat-top-actions"><button class="btn btn-chip sm" data-cc-side-toggle title="' + esc(t("\u4F1A\u8BDD\u6E05\u5355\u5F00\u5408\uFF08\u5177\u4F53\u4F1A\u8BDD\u6DF1\u94FE\u9ED8\u8BA4\u6536\u8D77\uFF09")) + '">' + esc(t("\u4F1A\u8BDD\u5217\u8868")) + '</button><button class="btn btn-chip sm" data-cc-provider title="' + esc(t("\u914D\u7F6E OpenAIProvider\uFF08api-key \u5F15\u7528 / base \u5730\u5740 / \u6A21\u578B\u5206\u7EA7\uFF09")) + '">Provider</button><span data-cc-state></span><button class="btn btn-chip sm" data-cc-new>' + esc(t("\u65B0\u4F1A\u8BDD")) + "</button></div></div>" + (can ? "" : '<div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u5BF9\u8BDD\u901A\u8DEF\u672A\u6CE8\u5165\uFF08chatfaces.js \u7F3A\u5E2D\uFF09\u2014\u2014\u8F6C\u5F55\u4E0E\u53D1\u9001\u9762\u663E\u5F0F\u7F3A\u5E2D\uFF0C\u4E0D\u5192\u5145\u3002")) + "</div></div></div>") + '<div data-sess-backlinks-slot class="chat-backlinks"></div><div class="chat-main"><aside class="chat-side"><div class="chat-side-head">' + esc(t("\u4F1A\u8BDD")) + '<span class="chat-side-new"></span></div><div class="chat-sess-list" data-cc-sessions><div class="chat-sess-day">' + esc(t("\u4F1A\u8BDD\u73B0\u8BFB\u4E2D\u2026")) + '</div></div></aside><div class="chat-stream-wrap"><div class="chat-stream" data-cc-tr><div class="chat-sess-day">' + esc(t("\u8F6C\u5F55\u73B0\u8BFB\u4E2D\u2026")) + '</div></div><button class="chat-jump" data-cc-jump hidden>\u2193 <span data-cc-jump-n></span></button></div><aside class="chat-right" data-cc-right><div data-stages-panel><div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u4E09\u6BB5\u5F0F Session \u73B0\u8BFB\u4E2D\u2026")) + '</div></div></div></div><div data-ctx-panel><div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u4E0A\u4E0B\u6587\u73B0\u8BFB\u4E2D\u2026\uFF08\u9009\u5B9A\u4F1A\u8BDD\u540E\u88C5\u914D\uFF09")) + '</div></div></div></div><div data-provider-drawer></div></aside></div><div class="chat-dock"><div class="chat-refs-slot" data-cc-refs></div>' + composerHtmlInner(can, false) + '<div class="chat-note" data-cc-note></div></div></div>';
  }
  function streamEl(el) {
    return el.querySelector("[data-cc-tr]");
  }
  function paintRefs(el) {
    CHAT.refs = { ...extractRefs(CHAT.lastNodes), chunks: CHAT.refsChunks };
    if (CHAT.paintRefs) CHAT.paintRefs(CHAT.refs);
  }
  async function refreshCtxChunks(el) {
    const cc = typeof window !== "undefined" && window.__CC_CONSOLE_CTX__ || null;
    if (!cc || !cc.api || typeof cc.api.sessionContext !== "function" || !CHAT.sessionId) {
      CHAT.refsChunks = [];
      paintRefs(el);
      return;
    }
    try {
      const r = await cc.api.sessionContext(CHAT.sessionId);
      const d = r && (r.ok || r.status === 200) && r.data && typeof r.data === "object" ? r.data : {};
      const pack = d.pack && typeof d.pack === "object" ? d.pack : {};
      CHAT.refsChunks = (Array.isArray(pack.chunks) ? pack.chunks : []).map((c) => ({ kind: String(c && c.kind || ""), scope: String(c && c.scope || "global"), name: String(c && c.name || ""), score: Number(c && c.score || 0) })).filter((c) => c.name);
    } catch {
      CHAT.refsChunks = [];
    }
    paintRefs(el);
  }
  function nearBottom(box) {
    return box.scrollHeight - box.scrollTop - box.clientHeight < 48;
  }
  function followBottom(box) {
    box.scrollTop = box.scrollHeight;
  }
  function applyIncremental(el, nodes) {
    const box = streamEl(el);
    if (!box) return { appended: 0, jumpVisible: false };
    let col = box.querySelector("[data-cc-col]");
    if (!col) {
      box.innerHTML = '<div class="chat-col" data-cc-col></div>';
      col = box.querySelector("[data-cc-col]");
    }
    const wasBottom = nearBottom(box);
    const fresh = nextBatch(nodes, CHAT.painted, CHAT.lastSeq === 0);
    if (CHAT.lastSeq === 0 && !CHAT.painted.size) {
      col.innerHTML = "";
    }
    for (const n of fresh) {
      CHAT.painted.add(nodeKey(n));
      if (Number(n.seq) && Number(n.seq) !== Number.MAX_SAFE_INTEGER && Number(n.seq) > CHAT.lastSeq) CHAT.lastSeq = Number(n.seq);
      col.insertAdjacentHTML("beforeend", nodeHtml(n, null));
    }
    classifyMentions(col);
    paintRefs(el);
    const pendBox = box.querySelector("[data-cc-pend]");
    const pendHtml = pendingNodes().map((n) => nodeHtml(n, null)).join("");
    if (pendHtml) {
      if (!pendBox) {
        col.insertAdjacentHTML("beforeend", "<div data-cc-pend>" + pendHtml + "</div>");
      } else {
        pendBox.innerHTML = pendHtml;
      }
    } else if (pendBox) {
      pendBox.remove();
    }
    const empty = box.querySelector("[data-cc-empty]");
    const isEmpty = !nodes.length && !CHAT.pending.length;
    if (isEmpty && !empty) col.insertAdjacentHTML("afterbegin", emptyStateHtml());
    else if (!isEmpty && empty) empty.remove();
    const jump = el.querySelector("[data-cc-jump]");
    if (wasBottom) {
      followBottom(box);
      if (jump) {
        jump.hidden = true;
      }
      return { appended: fresh.length, jumpVisible: false };
    }
    if (jump && fresh.length) {
      jump.hidden = false;
      const n = jump.querySelector("[data-cc-jump-n]");
      const prev = Number(n && n.textContent) || 0;
      if (n) n.textContent = String(prev + fresh.length);
    }
    return { appended: fresh.length, jumpVisible: !!(jump && !jump.hidden) };
  }
  async function refreshSessions(el, f) {
    const box = el.querySelector("[data-cc-sessions]");
    if (!box) return;
    const r = await f.sessions();
    if (r.status === 200) {
      box.innerHTML = sessionsHtml(r.data && r.data.items, CHAT.sessionId, CHAT.seenSeq);
    } else {
      box.innerHTML = '<div class="chat-sess-day">' + esc(t("\u4F1A\u8BDD\u6E05\u5355\u6682\u4E0D\u53EF\u8BFB\uFF08") + String(r.status) + t("\uFF09\u2014\u2014\u4E0D\u5192\u5145\u7A7A\u3002")) + "</div>";
    }
    box.querySelectorAll("[data-sess-pick]").forEach((row) => {
      row.addEventListener("click", () => {
        pickSession(el, f, String(row.getAttribute("data-sess-pick") || ""));
      });
    });
  }
  function pickSession(el, f, id) {
    if (!id || id === CHAT.sessionId) return;
    CHAT.sessionId = id;
    CHAT.lastSeq = 0;
    CHAT.painted = /* @__PURE__ */ new Set();
    CHAT.lastNodes = [];
    CHAT.refsChunks = [];
    CHAT.state = null;
    paintState(el);
    refreshSessions(el, f);
    refreshTranscript(el, f);
    ensureCtxBar(el, id);
    refreshCtxChunks(el);
  }
  function paintState(el) {
    const slot = el.querySelector("[data-cc-state]");
    if (slot) slot.innerHTML = stateChipsHtml(CHAT.state);
    const live = el.querySelector("[data-cc-live]");
    const st = CHAT.state;
    if (st && st.running === true) {
      const html3 = liveRowHtml(st);
      if (live) {
        live.innerHTML = html3;
      } else {
        const box = streamEl(el);
        if (box) {
          const col = box.querySelector("[data-cc-col]");
          (col || box).insertAdjacentHTML("afterbegin", "<div data-cc-live>" + html3 + "</div>");
        }
      }
    } else if (live) {
      live.remove();
    }
    const btn = el.querySelector("[data-cc-send]");
    const running = !!(st && st.running === true);
    if (btn && running !== (btn.dataset.mode === "steer")) {
      btn.dataset.mode = running ? "steer" : "queue";
      btn.title = running ? t("\u505C\u6B62") : t("\u53D1\u9001");
      const ico = btn.querySelector(".chat-send-ico");
      if (ico) ico.textContent = running ? "\u25A0" : "\u2191";
      btn.classList.toggle("stop", running);
    }
  }
  async function refreshState(el, f) {
    if (!CHAT.sessionId) {
      CHAT.state = null;
      paintState(el);
      return;
    }
    const r = await f.state(CHAT.sessionId);
    if (r.status !== 200 || !r.data || typeof r.data !== "object") {
      return;
    }
    CHAT.state = r.data;
    try {
      window.__CC_CHAT_STATE__ = CHAT.state;
      repaintSessionStages(el);
    } catch {
    }
    paintState(el);
  }
  async function refreshTranscript(el, f) {
    const box = streamEl(el);
    if (!box) return;
    if (!CHAT.sessionId) {
      box.innerHTML = '<div class="chat-col" data-cc-col>' + emptyStateHtml() + "</div>";
      return;
    }
    const r = await f.transcript(CHAT.sessionId, 80);
    if (r.status !== 200) {
      box.innerHTML = '<div class="chat-col" data-cc-col><div class="chat-empty"><div class="chat-empty-sub">' + esc(t("\u8F6C\u5F55\u9762\u6682\u4E0D\u53EF\u8BFB\uFF08") + String(r.status) + (r.data && r.data.error ? " \xB7 " + String(r.data.error) : "") + t("\uFF09\u2014\u2014\u4E0D\u5192\u5145\u7A7A\u3002")) + "</div></div></div>";
      CHAT.lastSeq = 0;
      CHAT.painted = /* @__PURE__ */ new Set();
      return;
    }
    const events = r.data && Array.isArray(r.data.events) ? r.data.events : [];
    if (CHAT.pending.length) {
      const byReq = new Set(events.map((e) => String(e && e.data && e.data.requestId || "")));
      const texts = new Set(events.filter((e) => e && e.type === "user/message").map((e) => String((e.data && e.data.content || []).filter((b) => b && b.type === "text").map((b) => b.text).join("\n"))));
      CHAT.pending = CHAT.pending.filter((p) => !(p.requestId && byReq.has(p.requestId)) && !texts.has(p.text));
    }
    const nodes = mapTranscript(events);
    CHAT.lastNodes = nodes;
    const lastEvSeq = events.reduce((m, e) => Math.max(m, Number(e && e.seq) || 0), 0);
    if (lastEvSeq) CHAT.seenSeq.set(CHAT.sessionId, Math.max(CHAT.seenSeq.get(CHAT.sessionId) || 0, lastEvSeq));
    applyIncremental(el, nodes);
    refreshState(el, f);
  }
  function ensureCtxBar(el, sessionId) {
    const rail = el.querySelector("[data-cc-right]");
    if (!rail) return;
    const ctx2 = typeof window !== "undefined" && window.__CC_CONSOLE_CTX__ || null;
    if (sessionId) {
      mountSessionCtx(el, ctx2, sessionId);
    }
    mountSessionStages(el);
    const provBtn = el.querySelector("[data-cc-provider]");
    if (provBtn && !provBtn.dataset.wired) {
      provBtn.dataset.wired = "1";
      provBtn.addEventListener("click", () => {
        const c = typeof window !== "undefined" && window.__CC_CONSOLE_CTX__ || null;
        if (c) openProviderDrawer(el, c, "openai-default");
      });
    }
  }
  async function doSend(el, f, textOverride) {
    const ta = el.querySelector("[data-cc-text]");
    const text2 = String(textOverride != null ? textOverride : ta ? ta.value : "").trim();
    if (!text2) return;
    if (!CHAT.sessionId) {
      const cr = await f.create();
      if (cr.status === 200 && cr.data && cr.data.sessionId) {
        CHAT.sessionId = String(cr.data.sessionId);
        refreshSessions(el, f);
        ensureCtxBar(el, CHAT.sessionId);
        refreshCtxChunks(el);
      } else {
        const note2 = el.querySelector("[data-cc-note]");
        if (note2) note2.textContent = t("\u65B0\u4F1A\u8BDD\u5931\u8D25\uFF08") + String(cr.status) + t("\uFF09");
        return;
      }
    }
    if (ta && textOverride == null) ta.value = "";
    autoGrow2(ta);
    const mode = modeForState(CHAT.state);
    const requestId = "cc-" + Date.now().toString(36) + "-" + Math.random().toString(36).slice(2, 8);
    CHAT.pending.push({ text: text2, requestId, at: Date.now() });
    applyIncremental(el, []);
    let r;
    try {
      r = await (mode === "steer" && typeof f.send === "function" ? f.send(CHAT.sessionId, text2, "steer") : f.send(CHAT.sessionId, text2));
    } catch (e) {
      r = { status: 0, data: { error: String(e && e.message || e) } };
    }
    if (!(r && r.status === 200)) {
      failPending(requestId, String(r && r.status) + (r && r.data && r.data.error ? " \xB7 " + String(r.data.error) : ""));
      applyIncremental(el, []);
    }
    setTimeout(() => refreshTranscript(el, f), 900);
    setTimeout(() => refreshTranscript(el, f), 3e3);
  }
  function failPending(requestId, errText) {
    const p = CHAT.pending.find((x) => x.requestId === requestId);
    if (p) CHAT.pending = CHAT.pending.filter((x) => x !== p);
    CHAT.failed = CHAT.failed || [];
    CHAT.failed.push({ text: p ? p.text : "", requestId, error: errText || "0" });
  }
  function bindRetry2(el, f) {
    el.addEventListener("click", (ev) => {
      const btn = ev.target.closest("[data-cc-retry]");
      if (!btn) return;
      const requestId = String(btn.getAttribute("data-cc-retry") || "");
      const p = (CHAT.failed || []).find((x) => x.requestId === requestId);
      if (!p) return;
      CHAT.failed = (CHAT.failed || []).filter((x) => x !== p);
      doSend(el, f, p.text);
    });
  }
  function bindImageUpload(el, f) {
    if (!f || typeof f.uploadImage !== "function") return;
    const input = el.querySelector("[data-cc-img]");
    const previews = el.querySelector("[data-cc-img-previews]");
    const ta = el.querySelector("[data-cc-text]");
    if (!input || !ta) return;
    const note2 = (msg) => {
      const n = el.querySelector("[data-cc-note]");
      if (n) n.textContent = msg;
    };
    const uploadOne = async (file) => {
      if (!file || !String(file.type || "").startsWith("image/")) return;
      if (file.size > 8 << 20) {
        note2(t("\u56FE\u7247\u8D85\u8FC7 8MB \u4E0A\u9650"));
        return;
      }
      note2(t("\u56FE\u7247\u4E0A\u4F20\u4E2D\u2026"));
      try {
        const url = await f.uploadImage(file);
        const tag2 = "![image](" + url + ")";
        ta.value = ta.value.trimEnd() + (ta.value.trimEnd() ? "\n" : "") + tag2;
        autoGrow2(ta);
        note2(t("\u56FE\u7247\u5DF2\u9644\u52A0\uFF08\u968F\u6D88\u606F\u53D1\u9001\uFF09"));
        if (previews) {
          const img = document.createElement("img");
          img.src = url;
          img.alt = "preview";
          img.style.cssText = "width:44px;height:44px;object-fit:cover;border-radius:6px;margin-left:6px";
          previews.appendChild(img);
        }
      } catch (e) {
        note2(t("\u56FE\u7247\u4E0A\u4F20\u5931\u8D25\uFF08") + String(e && e.message || e) + t("\uFF09"));
      }
    };
    input.addEventListener("change", () => {
      Array.from(input.files || []).forEach(uploadOne);
      input.value = "";
    });
    ta.addEventListener("paste", (ev) => {
      const items = Array.from(ev.clipboardData && ev.clipboardData.items || []);
      const hit = items.find((it) => String(it.type || "").startsWith("image/"));
      if (!hit) return;
      const file = hit.getAsFile();
      if (file) {
        ev.preventDefault();
        uploadOne(file);
      }
    });
    if (previews) previews.addEventListener("click", (ev) => {
      const img = ev.target.closest("img");
      if (!img) return;
      const src = img.getAttribute("src") || "";
      ta.value = ta.value.split("\n").filter((ln) => ln.trim() !== "![image](" + src + ")").join("\n");
      autoGrow2(ta);
      img.remove();
    });
  }
  async function doStop(el, f) {
    const note2 = el.querySelector("[data-cc-note]");
    if (!CHAT.sessionId) return;
    if (typeof f.stop !== "function") {
      if (note2) note2.textContent = t("\u505C\u6B62\u901A\u8DEF\u672A\u6CE8\u5165\uFF08chatfaces.js.stop \u7F3A\u5E2D\uFF09\u2014\u2014\u663E\u5F0F\u7F3A\u5E2D\uFF0C\u4E0D\u5192\u5145\u3002");
      return;
    }
    let r;
    try {
      r = await f.stop(CHAT.sessionId);
    } catch (e) {
      r = { status: 0, data: { error: String(e && e.message || e) } };
    }
    if (r && r.status === 200) {
      if (note2) note2.textContent = t("\u5DF2\u8BF7\u6C42\u505C\u6B62\u2014\u2014\u4F1A\u8BDD\u56DE\u5230\u7A7A\u95F2\u540E\u8F6C\u5F55\u53EF\u89C1\u4E2D\u65AD\u8BCD\u5F62\u3002");
    } else {
      if (note2) note2.textContent = t("\u505C\u6B62\u5931\u8D25\uFF08") + String(r && r.status) + (r && r.data && r.data.error ? " \xB7 " + String(r.data.error) : "") + t("\uFF09");
    }
    setTimeout(() => refreshState(el, f), 800);
  }
  function autoGrow2(ta) {
    if (!ta) return;
    ta.style.height = "auto";
    const lh = 24;
    const max = lh * 8 + 8;
    ta.style.height = Math.min(ta.scrollHeight, max) + "px";
  }
  function defaultSideOpen(id) {
    return !(id != null && id !== "");
  }
  function paintSideToggle(el, open) {
    const aside = el.querySelector(".chat-side");
    if (aside) aside.hidden = !open;
    const btn = el.querySelector("[data-cc-side-toggle]");
    if (btn) {
      btn.classList.toggle("acc", open);
      btn.setAttribute("aria-pressed", open ? "true" : "false");
    }
  }
  function dialogScreen(el, ctx2, sideOpen) {
    const f = faces3();
    const can = !!(f && typeof f.send === "function");
    CHAT.el = el;
    el.innerHTML = screenShell(can);
    el.classList.add("route-chat");
    paintSideToggle(el, sideOpen !== false);
    el.querySelector("[data-cc-side-toggle]").addEventListener("click", () => {
      const aside = el.querySelector(".chat-side");
      paintSideToggle(el, !(aside && !aside.hidden));
    });
    try {
      const pre = String(typeof window !== "undefined" && window.__CC_CHAT_PREFILL__ || "");
      if (pre) {
        const ta = el.querySelector("[data-cc-text]");
        if (ta && !ta.value) {
          ta.value = pre;
          ta.dispatchEvent(new Event("input", { bubbles: true }));
          ta.focus();
        }
        delete window.__CC_CHAT_PREFILL__;
      }
    } catch {
    }
    const paintLinks = () => {
      const slot = el.querySelector("[data-sess-backlinks-slot]");
      if (slot) slot.innerHTML = sessBacklinksHtml(ctx2.store, CHAT.sessionId);
      slot && slot.querySelectorAll("[data-go]").forEach((n) => {
        n.addEventListener("click", () => ctx2.go(n.getAttribute("data-go")));
      });
    };
    paintLinks();
    const unsubLinks = ctx2.store.subscribe(() => {
      paintLinks();
      refreshSessionsSoon(el, f);
    });
    el.addEventListener("disconnected", () => {
      try {
        unsubLinks();
      } catch {
      }
      CHAT.el = null;
    }, { once: true });
    const newBtn = el.querySelector("[data-cc-new]");
    if (newBtn) newBtn.addEventListener("click", async () => {
      newBtn.disabled = true;
      const r = await f.create();
      newBtn.disabled = false;
      if (r.status === 200 && r.data && r.data.sessionId) {
        CHAT.sessionId = String(r.data.sessionId);
        CHAT.lastSeq = 0;
        CHAT.painted = /* @__PURE__ */ new Set();
        CHAT.lastNodes = [];
        CHAT.refsChunks = [];
        refreshSessions(el, f);
        refreshTranscript(el, f);
        ensureCtxBar(el, CHAT.sessionId);
        refreshCtxChunks(el);
      } else {
        const note2 = el.querySelector("[data-cc-note]");
        if (note2) note2.textContent = t("\u65B0\u4F1A\u8BDD\u5931\u8D25\uFF08") + String(r.status) + t("\uFF09");
      }
    });
    const sendBtn = el.querySelector("[data-cc-send]");
    const textTa = el.querySelector("[data-cc-text]");
    if (sendBtn && textTa) {
      sendBtn.addEventListener("click", () => {
        if (sendBtn.dataset.mode === "steer" && CHAT.state && CHAT.state.running === true) {
          doStop(el, f);
        } else {
          doSend(el, f);
        }
      });
      textTa.addEventListener("keydown", (ev) => {
        if (ev.key === "Enter" && !ev.shiftKey) {
          ev.preventDefault();
          doSend(el, f);
        }
      });
      textTa.addEventListener("input", () => autoGrow2(textTa));
    }
    el.addEventListener("click", (ev) => {
      const s = ev.target.closest("[data-cc-sample]");
      if (!s) return;
      const ta = el.querySelector("[data-cc-text]");
      if (ta) {
        ta.value = s.textContent.trim();
        autoGrow2(ta);
        ta.focus();
      }
    });
    const jump = el.querySelector("[data-cc-jump]");
    if (jump) jump.addEventListener("click", () => {
      const box = streamEl(el);
      if (box) followBottom(box);
      jump.hidden = true;
      const n = jump.querySelector("[data-cc-jump-n]");
      if (n) n.textContent = "";
    });
    bindRetry2(el, f);
    bindImageUpload(el, f);
    const cc = typeof window !== "undefined" && window.__CC_CONSOLE_CTX__ || null;
    if (cc && cc.api) {
      CHAT.paintRefs = mountRefs(el, { api: cc.api, go: (x) => ctx2.go(x) });
    }
    el.addEventListener("click", (ev) => {
      const copy = ev.target.closest(".md-code-copy");
      if (copy) {
        const block2 = copy.closest(".md-code-block");
        const code = block2 ? block2.querySelector("pre code") : null;
        const text2 = code ? String(code.textContent || "") : "";
        if (text2 && navigator.clipboard && typeof navigator.clipboard.writeText === "function") {
          navigator.clipboard.writeText(text2).then(() => {
            copy.textContent = t("\u590D\u5236\u6210\u529F");
            copy.classList.add("ok");
            setTimeout(() => {
              copy.textContent = t("\u590D\u5236");
              copy.classList.remove("ok");
            }, 1e3);
          }).catch(() => {
          });
        }
        return;
      }
      const wrap = ev.target.closest(".md-code-wrap");
      if (wrap) {
        const block2 = wrap.closest(".md-code-block");
        if (block2) {
          const on = block2.classList.toggle("wrap");
          wrap.textContent = on ? t("\u81EA\u52A8\u6362\u884C") : t("\u53D6\u6D88\u81EA\u52A8\u6362\u884C");
        }
      }
    });
    ensureCtxBar(el, CHAT.sessionId);
    refreshCtxChunks(el);
    refreshSessions(el, f);
    refreshTranscript(el, f);
    if (CHAT.timer) clearInterval(CHAT.timer);
    CHAT.timer = setInterval(() => {
      const root = CHAT.el && CHAT.el.isConnected ? CHAT.el : document.querySelector(".route-chat");
      if (!root || !root.isConnected) {
        clearInterval(CHAT.timer);
        CHAT.timer = 0;
        return;
      }
      const f2 = faces3();
      if (!f2) return;
      refreshTranscript(root, f2);
      refreshSessionsSoon(root, f2);
    }, 3500);
  }
  var sessTick = 0;
  function refreshSessionsSoon(el, f) {
    sessTick++;
    if (sessTick % 4 !== 1) return;
    refreshSessions(el, f);
  }
  function render9(el, ctx2) {
    const id = ctx2.params && ctx2.params.id;
    const agent = ctx2.params && ctx2.params.agent;
    const paramKey = id != null && id !== "" ? "session/" + String(id) : "chat/" + String(agent || "");
    const fresh = LAST.route !== String(ctx2.route || "") || LAST.paramKey !== paramKey || LAST.locale !== getLocale();
    LAST.route = String(ctx2.route || "");
    LAST.paramKey = paramKey;
    LAST.locale = getLocale();
    if (fresh) {
      if (id != null && id !== "") CHAT.sessionId = String(id);
      else CHAT.sessionId = "";
      CHAT.pending = [];
      CHAT.failed = [];
      CHAT.lastSeq = 0;
      CHAT.painted = /* @__PURE__ */ new Set();
      CHAT.lastNodes = [];
      CHAT.refsChunks = [];
      CHAT.paintRefs = null;
      dialogScreen(el, ctx2, defaultSideOpen(id));
      if (id == null && agent) {
        const f2 = faces3();
        if (f2) setTimeout(async () => {
          if (CHAT.sessionId) return;
          const r = await f2.sessions();
          const items = r.status === 200 && r.data && Array.isArray(r.data.items) ? r.data.items : [];
          const pick = items.filter((s) => s && s.sessionId).sort((a, b) => (Number(b.updatedAt) || 0) - (Number(a.updatedAt) || 0))[0];
          if (pick && !CHAT.sessionId && CHAT.el === el) {
            pickSession(el, f2, String(pick.sessionId));
          }
        }, 700);
      }
      return;
    }
    const f = faces3();
    if (f) refreshTranscript(el, f);
  }

  // src/screens/tags.js
  function frontmatterTags(content) {
    const s = String(content || "");
    const head = s.slice(0, 4e3);
    if (!head.includes("tags:")) return [];
    const out = [];
    const flat = head.match(/tags:\s*\[([^\]]*)\]/);
    if (flat) {
      flat[1].split(",").forEach((x) => {
        const v = x.trim().replace(/^['"]|['"]$/g, "");
        if (v) out.push(v);
      });
      return out;
    }
    const block2 = head.match(/tags:\s*\n((?:[ \t]+-[ \t]+[^\n]+\n?)+)/);
    if (block2) {
      block2[1].split("\n").forEach((line) => {
        const m = /^[ \t]+-[ \t]+(.+?)\s*$/.exec(line);
        if (m) {
          const v = m[1].trim().replace(/^['"]|['"]$/g, "");
          if (v) out.push(v);
        }
      });
    }
    return out;
  }
  function tagsOfRef(o) {
    if (!o || typeof o !== "object") return [];
    const raw = o.raw && typeof o.raw === "object" ? o.raw : o;
    const spec = raw.spec && typeof raw.spec === "object" ? raw.spec : {};
    const meta = spec.metadata && typeof spec.metadata === "object" ? spec.metadata : raw.metadata && typeof raw.metadata === "object" ? raw.metadata : {};
    const list2 = meta.tags || spec.tags || raw.tags || o.tags;
    return Array.isArray(list2) ? list2.map((x) => String(x || "").trim()).filter(Boolean) : [];
  }
  function docKeyOf(kind, scope, name) {
    return String(kind || "") + "|" + String(scope || "global") + "|" + String(name || "");
  }
  function aggregateTags(indexDocs, objs, cached, vocab) {
    const V = vocab || {};
    const UN = t("\u672A\u5206\u7EC4");
    const m = /* @__PURE__ */ new Map();
    const add = (tag2, key) => {
      const t2 = String(tag2 || "").trim();
      if (!t2 || !key) return;
      let set2 = m.get(t2);
      if (!set2) {
        set2 = /* @__PURE__ */ new Set();
        m.set(t2, set2);
      }
      set2.add(key);
    };
    (Array.isArray(indexDocs) ? indexDocs : []).forEach((d) => {
      if (!d || !d.name) return;
      const key = docKeyOf(d.kind, d.scope, d.name);
      const tags = Array.isArray(d.tags) ? d.tags : [];
      tags.forEach((t2) => add(t2, key));
    });
    (Array.isArray(objs) ? objs : []).forEach((o) => {
      if (!o || !o.kind || !o.name) return;
      const key = docKeyOf(o.kind, o.scope, o.name);
      tagsOfRef(o).forEach((t2) => add(t2, key));
      const c = typeof cached === "function" ? cached(o.kind, o.scope || "global", o.name, o.rv) : null;
      if (c && typeof c.content === "string") {
        frontmatterTags(c.content).forEach((t2) => add(t2, key));
      }
    });
    const out = Array.from(m.entries()).map(([tag2, set2]) => {
      const colon = tag2.indexOf(":");
      const prefix = colon > 0 ? tag2.slice(0, colon) : "";
      const type2 = V[tag2] || (prefix ? V[prefix] || prefix : "") || UN;
      return { tag: tag2, type: type2, n: set2.size, keys: Array.from(set2).sort() };
    });
    out.sort((a, b) => b.n - a.n || (a.tag < b.tag ? -1 : 1));
    return out;
  }
  function groupByType(rows) {
    const UN = t("\u672A\u5206\u7EC4");
    const m = /* @__PURE__ */ new Map();
    (rows || []).forEach((r) => {
      const key = r && r.type ? r.type : UN;
      if (!m.has(key)) m.set(key, []);
      m.get(key).push(r);
    });
    return Array.from(m.keys()).sort((a, b) => a === UN ? 1 : b === UN ? -1 : a < b ? -1 : 1).map((type2) => ({ type: type2, rows: m.get(type2) }));
  }
  var INDEX_CACHE = { docs: null, at: 0 };
  var INDEX_TTL_MS = 3e4;
  async function loadContextIndexDocs(api2) {
    if (INDEX_CACHE.docs && Date.now() - INDEX_CACHE.at < INDEX_TTL_MS) {
      return { ok: true, docs: INDEX_CACHE.docs };
    }
    const fn = api2 && (typeof api2.contextIndex === "function" ? () => api2.contextIndex() : () => api2.request("v1/context/index"));
    if (!fn) return { ok: false, docs: [], error: "api \u7F3A\u5E2D" };
    const r = await fn();
    if (!r || !r.ok) return { ok: false, docs: [], error: String(r && (r.error || r.status) || "\u4E0D\u53EF\u8BFB") };
    const docs = r.data && Array.isArray(r.data.docs) ? r.data.docs : [];
    INDEX_CACHE = { docs, at: Date.now() };
    return { ok: true, docs };
  }
  function hasCachedIndex() {
    return !!(INDEX_CACHE.docs && Date.now() - INDEX_CACHE.at < INDEX_TTL_MS);
  }
  function cachedIndexTagsOf(kind, scope, name) {
    if (!INDEX_CACHE.docs) return [];
    const k = docKeyOf(kind, scope, name);
    const hit = INDEX_CACHE.docs.find((d) => d && docKeyOf(d.kind, d.scope, d.name) === k);
    return hit && Array.isArray(hit.tags) ? hit.tags : [];
  }
  function render10(el, ctx2) {
    const store2 = ctx2.store;
    let vocab = tagVocabFromKinds(store2.kinds());
    el.innerHTML = '<div class="view-inner">' + screenHead(
      t("\u6807\u7B7E"),
      t("tagType \u5206\u7EC4\u805A\u5408\uFF08\u4E09\u6E90\u4E25\u683C\u53BB\u91CD\uFF1Acontext/index \u7D22\u5F15\u6295\u5F71 \u222A \u5BF9\u8C61 tags \u222A \u7F13\u5B58 frontmatter\uFF1B\u96F6\u5047\u8BCD\u8868\uFF09\u2014\u2014\u70B9 tag \u770B\u547D\u4E2D\u6587\u6863"),
      chipGo(t("\u6587\u6863\u5E93"), "docs", "mute", false)
    ) + '<div data-tags-body><div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u805A\u5408\u73B0\u8BFB\u4E2D\u2026\uFF08GET /v1/context/index + \u7F13\u5B58\u6B63\u6587 frontmatter\uFF09")) + "</div></div></div></div></div>";
    const paint3 = (rows, sourceNote) => {
      const body = el.querySelector("[data-tags-body]");
      if (!body) return;
      if (!rows.length) {
        body.innerHTML = '<div class="card"><div class="card-b">' + emptyBox(t("\u6682\u65E0\u6807\u7B7E"), t("\u4E09\u6E90\u805A\u5408\u96F6\u547D\u4E2D\uFF08\u7D22\u5F15\u6295\u5F71\u4E0E\u9884\u53D6\u7A97\u53E3\u6B63\u6587\u5747\u672A\u58F0\u660E tags\uFF09\u3002")) + "</div></div>";
        return;
      }
      const groups = groupByType(rows);
      body.innerHTML = '<div class="small faint" style="margin-bottom:8px">' + esc(sourceNote) + "</div>" + groups.map((g) => '<section class="card" style="margin-bottom:10px"><div class="card-h"><span class="h-title">' + esc(g.type) + '</span><span class="h-right">' + chip(g.rows.length + t(" \u4E2A tag"), "mute") + '</span></div><div class="card-b" style="display:flex;gap:6px;flex-wrap:wrap">' + g.rows.map((r) => chipGo("#" + r.tag + " \xB7 " + r.n, tagRoute(r.tag), r.n >= 2 ? "acc" : "mute", true, t("\u770B\u540C tag \u6587\u6863"))).join("") + "</div></section>").join("");
    };
    const cachedFn = (k, s, n, rv) => store2.cachedDoc(k, s, n, rv);
    paint3(
      aggregateTags(INDEX_CACHE.docs, store2.objects(), cachedFn, vocab),
      INDEX_CACHE.docs ? t("\u6570\u636E\u6E90\uFF1AGET /v1/context/index\uFF08") + INDEX_CACHE.docs.length + t(" \u7BC7\uFF09\u222A \u5BF9\u8C61 tags \u222A \u7F13\u5B58 frontmatter\u2014\u2014\u53BB\u91CD\u8BA1\u6570\u3002") : t("\u6570\u636E\u6E90\uFF1A\u5BF9\u8C61 tags \u222A \u7F13\u5B58 frontmatter\uFF08\u9884\u53D6\u7A97\u53E3 24 \u7BC7\uFF09\u2014\u2014\u7D22\u5F15\u6295\u5F71\u73B0\u8BFB\u4E2D\u2026")
    );
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const stale = store2.staleDocs ? store2.staleDocs(24) : [];
    const hydrateP = stale.length && ctx2.api && typeof ctx2.api.object === "function" ? (async () => {
      let i = 0;
      const w = async () => {
        while (i < stale.length) {
          const r = stale[i++];
          const res = await ctx2.api.object(r.kind, r.scope || "global", r.name);
          if (res && res.ok) {
            const d = res.data && typeof res.data === "object" ? res.data : {};
            const c = d.content != null ? String(d.content) : "";
            if (c) store2.cacheDoc(r.kind, r.scope || "global", r.name, r.rv, c);
          }
        }
      };
      await Promise.all(Array.from({ length: Math.min(4, stale.length) }, w));
    })() : Promise.resolve();
    const kindsP = ctx2.api && typeof ctx2.api.kinds === "function" ? ctx2.api.kinds().catch(() => null) : Promise.resolve(null);
    Promise.all([loadContextIndexDocs(ctx2.api), hydrateP, kindsP]).then(([res, , kr]) => {
      if (!el.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
      if (kr && kr.ok && kr.data && Array.isArray(kr.data.kinds)) {
        vocab = tagVocabFromKinds(kr.data.kinds);
      }
      if (!res.ok) {
        paint3(
          aggregateTags(null, store2.objects(), cachedFn, vocab),
          t("\u7D22\u5F15\u6295\u5F71\u6682\u4E0D\u53EF\u8BFB\uFF08") + res.error + t("\uFF09\u2014\u2014\u5F53\u524D\u4EC5\u5C55\u793A\u5BF9\u8C61 tags \u222A \u9884\u53D6\u7A97\u53E3 frontmatter\uFF08\u8BDA\u5B9E\u964D\u7EA7\uFF09\u3002")
        );
        return;
      }
      paint3(
        aggregateTags(res.docs, store2.objects(), cachedFn, vocab),
        t("\u6570\u636E\u6E90\uFF1AGET /v1/context/index \u7D22\u5F15\u6295\u5F71\uFF08") + res.docs.length + t(" \u7BC7\u5728\u518C\uFF09\u222A \u5BF9\u8C61 tags \u222A \u7F13\u5B58 frontmatter\uFF08\u9884\u53D6\u7A97\u53E3\uFF09\u2014\u2014\u4E09\u6E90\u5BF9\u8C61\u952E\u53BB\u91CD\u3002")
      );
    }).catch(() => {
    });
  }

  // src/screens/misc.js
  var treeOpen = /* @__PURE__ */ new Set();
  var treeSeeded = false;
  var docFilter = { kind: "", q: "", tag: "", routeTag: "" };
  var propOpen = /* @__PURE__ */ new Set();
  var dd = { tree: null, sel: null, cache: /* @__PURE__ */ new Map() };
  function treeLandingRoute(o) {
    if (o && o.kind === "ServiceDesc" && o.name) {
      return "designdoc/" + encodeURIComponent(repoOfScope(o.scope || o.name)) + "/" + encodeURIComponent("goal.md");
    }
    return objRoute(o.kind, o.scope, o.name);
  }
  function repoOfScope(scope) {
    const s = String(scope || "").toLowerCase();
    if (s.includes("instance")) return "core_system_instance";
    if (s.includes("workservice") || s.includes("deploy") || s.includes("storage") || s.includes("user")) return "core_design";
    return "core_design";
  }
  function treeScreen(el, ctx2) {
    const objs = ctx2.store.objects();
    const kinds = Array.from(new Set(objs.map((o) => o.kind))).sort();
    const scopeMap = {};
    objs.forEach((o) => {
      const k = scopeMap[o.kind] = scopeMap[o.kind] || {};
      (k[o.scope] = k[o.scope] || []).push(o);
    });
    if (!treeSeeded) {
      kinds.forEach((kind) => {
        if (objs.filter((o) => o.kind === kind).length <= 3) treeOpen.add(kind);
      });
      treeSeeded = true;
    }
    const body = kinds.map((kind) => {
      const open = treeOpen.has(kind);
      const n = objs.filter((o) => o.kind === kind).length;
      const rows = open ? Object.keys(scopeMap[kind] || {}).sort().map((scope) => {
        const items = (scopeMap[kind][scope] || []).slice().sort((a, b) => a.name < b.name ? -1 : 1);
        return '<div class="tree-scope"><span class="small faint mono">' + esc(scope) + "/</span>" + items.map((o) => '<div class="tree-row leaf" data-go="' + esc(treeLandingRoute(o)) + '"><span style="flex:1;min-width:0">' + esc(o.name) + '</span> <span class="faint tiny mono">r' + esc(String(o.rv || 0)) + "</span>" + (o.kind === "ServiceDesc" ? '<span class="ws-chip">' + chipGo(t("Service \u5927\u5C4F"), "service/" + encodeURIComponent(String(o.name || "")), "acc", true, t("\u6DF1\u94FE Service \u5927\u5C4F")) + "</span>" : "") + "</div>").join("") + "</div>";
      }).join("") : "";
      return '<div class="tree-node"><div class="tree-row grp" data-tree-toggle="' + esc(kind) + '"><b>' + esc(kind) + "</b>" + chip(n + t(" \u4E2A"), n ? "acc" : "mute") + '<span class="faint">' + (open ? "\u25BE" : "\u25B8") + "</span></div>" + rows + "</div>";
    }).join("");
    el.innerHTML = '<div class="view-inner">' + screenHead(t("\u7CFB\u7EDF\u6811"), t("kinds \u2192 scope \u2192 \u5BF9\u8C61\uFF08\u5FEB\u7167\u805A\u5408\u6295\u5F71\uFF1B\u5BF9\u8C61\u540D\u53EF\u70B9\u8FDB\u5DE5\u4F5C\u53F0\uFF1B\u22643 \u5BF9\u8C61\u7684 kind \u9ED8\u8BA4\u5C55\u5F00\uFF09")) + (kinds.length ? '<div class="tree-box" style="min-width:min(680px,100%)">' + body + "</div>" : emptyBox(t("\u6682\u65E0\u5BF9\u8C61"), t("\u7B49\u5F85\u9996\u4E2A\u5FEB\u7167\u805A\u5408\u3002"))) + "</div>";
    el.querySelectorAll("[data-tree-toggle]").forEach((row) => {
      row.addEventListener("click", () => {
        const k = row.dataset.treeToggle;
        if (treeOpen.has(k)) treeOpen.delete(k);
        else treeOpen.add(k);
        treeScreen(el, ctx2);
      });
    });
  }
  function fmMeta(content) {
    if (typeof content !== "string") return null;
    const tags = frontmatterTags(content);
    const head = content.slice(0, 4e3);
    const imp = head.match(/^importance:\s*(\d{1,3})\s*$/m);
    const importance = imp ? Number(imp[1]) : null;
    if (!tags.length && importance == null) return null;
    return { tags, importance };
  }
  function objectTagsOf(ctx2, o) {
    const set2 = /* @__PURE__ */ new Set();
    cachedIndexTagsOf(o.kind, o.scope || "global", o.name).forEach((x) => set2.add(x));
    tagsOfRef(o).forEach((x) => set2.add(x));
    const c = ctx2.store.cachedDoc(o.kind, o.scope || "global", o.name, o.rv);
    if (c) frontmatterTags(c.content).forEach((x) => set2.add(x));
    return Array.from(set2);
  }
  var docsIdxLoading = false;
  function docsScreen(el, ctx2) {
    const routeTag = String(ctx2.params && ctx2.params.tag || "");
    if (routeTag) {
      if (docFilter.routeTag !== routeTag) {
        docFilter.tag = routeTag;
        docFilter.routeTag = routeTag;
      }
    } else if (docFilter.routeTag) {
      docFilter.tag = "";
      docFilter.routeTag = "";
    }
    let objs = ctx2.store.objects();
    const presentKinds = Array.from(new Set(objs.map((o) => o.kind)));
    const DOC_FIRST = ["Goal", "Rule", "Runbook", "DesignDoc", "DriftReport", "DocReview", "BugDoc"];
    const kindChips = presentKinds.slice().sort((a, b) => {
      const ia = DOC_FIRST.indexOf(a), ib = DOC_FIRST.indexOf(b);
      return (ia < 0 ? DOC_FIRST.length + 1 : ia) - (ib < 0 ? DOC_FIRST.length + 1 : ib) || (a < b ? -1 : 1);
    });
    const pool = objs;
    let rows = pool;
    if (docFilter.kind) rows = rows.filter((o) => o.kind === docFilter.kind);
    if (docFilter.q) {
      const q = docFilter.q.toLowerCase();
      rows = rows.filter((o) => (o.kind + "/" + o.scope + "/" + o.name).toLowerCase().includes(q));
    }
    if (docFilter.tag) {
      const tg = docFilter.tag.toLowerCase();
      rows = rows.filter((o) => objectTagsOf(ctx2, o).some((x) => x.toLowerCase().includes(tg)));
      if (!hasCachedIndex() && !docsIdxLoading && ctx2.api) {
        docsIdxLoading = true;
        const stamp = typeof location !== "undefined" ? location.hash : "";
        loadContextIndexDocs(ctx2.api).then((res) => {
          docsIdxLoading = false;
          if (res && res.ok && el.isConnected && (typeof location !== "undefined" ? location.hash : "") === stamp) {
            docsScreen(el, ctx2);
          }
        }).catch(() => {
          docsIdxLoading = false;
        });
      }
    }
    const draftSet = new Set(ctx2.store.drafts().map((d) => d.kind + "|" + (d.scope || "global") + "|" + d.name));
    el.innerHTML = '<div class="view-inner">' + screenHead(
      t("\u6587\u6863\u5E93"),
      t("\u5168\u90E8\u6587\u6863\u4E0E\u5BF9\u8C61\u2014\u2014\u6309\u6807\u7B7E/\u7C7B\u578B\u8FC7\u6EE4\uFF1B\u70B9\u884C\u8FDB\u6E32\u67D3\u89C6\u56FE"),
      chipGo(t("\u6807\u7B7E\u805A\u5408"), "tags", "mute", false)
    ) + '<div class="card" style="margin-bottom:10px"><div class="card-b small">' + esc(t("\u65B0\u5EFA\u6587\u6863\uFF1A\u5BF9\u8BDD\u63D0\u4EA4\u9700\u6C42\uFF08\u603B\u89C8\u6216\u4F1A\u8BDD\u5C4F\uFF09\u2192 agent \u8D77\u8349 \u2192 Apply \u843D\u5E93\u2014\u2014\u6587\u6863\u521B\u5EFA\u8D70\u5BF9\u8BDD+Apply\uFF0C\u65E0\u72EC\u7ACB\u8868\u5355\u3002")) + " " + chipGo(t("\u6253\u5F00\u5BF9\u8BDD"), "chat/core-headless", "acc", true, t("\u5728\u4F1A\u8BDD\u5C4F\u63D0\u4EA4\u9700\u6C42\u2014\u2014agent \u8D77\u8349\u540E Apply \u843D\u5E93")) + (docFilter.tag ? ' <span class="chip warn tiny click" data-clear-tag="1" title="' + esc(t("\u6E05\u9664 tag \u8FC7\u6EE4")) + '">' + esc(t("\u6E05\u9664 tag\uFF1A")) + esc(docFilter.tag) + " \xD7</span>" : "") + (routeTag ? '<div class="small faint" style="margin-top:6px">' + esc(t("\u5F53\u524D URL \u76F4\u94FE tag \u8FC7\u6EE4\uFF08#/docs/tag/")) + esc(routeTag) + esc(t("\uFF09\u2014\u2014\u4E09\u6E90\u771F\u5B9E\u547D\u4E2D\uFF0C\u53EF\u5237\u65B0\u6062\u590D\u3002")) + "</div>" : "") + '</div></div><div class="filter-bar">' + t('<input class="inp" id="docs-q" placeholder="\u6309 kind/scope/name \u8FC7\u6EE4\u2026" value="') + esc(docFilter.q) + '">' + t('<input class="inp" id="docs-tag" placeholder="tag \u8FC7\u6EE4\uFF08\u4E09\u6E90\u805A\u5408\u73B0\u5339\u914D\uFF09\u2026" value="') + esc(docFilter.tag || "") + '" style="max-width:220px"><span class="fchip' + (!docFilter.kind ? " on" : "") + t('" data-docs-kind="">\u5168\u90E8</span>') + kindChips.map((k) => '<span class="fchip' + (docFilter.kind === k ? " on" : "") + '" data-docs-kind="' + esc(k) + '">' + esc(k) + "</span>").join("") + "</div>" + (rows.length ? '<div class="card">' + rows.slice(0, 300).map((o) => {
      const c = ctx2.store.cachedDoc(o.kind, o.scope || "global", o.name, o.rv);
      const meta = c ? fmMeta(c.content) : null;
      const allTags = objectTagsOf(ctx2, o);
      const tagChips = allTags.slice(0, 4).map((x) => chipGo("#" + x, tagRoute(x), "mute", true, t("\u770B\u540C tag \u6587\u6863"))).join(" ");
      const imp = meta && meta.importance != null ? chip("I:" + meta.importance, meta.importance >= 80 ? "acc" : "mute") : chip(t("\u91CD\u8981\u5EA6\u2014"), "mute");
      const sync = syncBadge(ctx2.store.runs(), o.name);
      const scope = o.scope || "global";
      const hasDraft = draftSet.has(o.kind + "|" + scope + "|" + o.name);
      const draftDot = hasDraft ? ' <span class="draft-dot click" data-go="approvals/' + encodeURIComponent(o.kind + "/" + scope + "/" + o.name) + '" title="' + esc(t("\u6709\u672A\u843D\u5E93\u8349\u7A3F\u2014\u2014\u70B9\u8FDB\u5BA1\u6279 diff \u5361")) + '">\u25CF ' + esc(t("\u8349\u7A3F\u5F85\u5BA1")) + "</span>" : "";
      return '<div class="mini-row" data-doc-row="' + esc(docRoute(o.kind, scope, o.name)) + '"><div class="mr-main"><div><b>' + esc(o.name) + "</b>" + draftDot + '</div><div class="mr-sub small mono">' + esc(o.kind + "/" + scope) + " \xB7 rv " + esc(String(o.rv != null ? o.rv : "\u2014")) + '</div><div class="mr-sub" style="display:flex;gap:4px;flex-wrap:wrap;align-items:center">' + tagChips + " " + imp + " " + sync + '</div></div><span class="ws-chip">' + chipGo(t("\u6253\u5F00"), docRoute(o.kind, scope, o.name), "acc", true, t("\u6E32\u67D3\u89C6\u56FE")) + "</span></div>";
    }).join("") + (rows.length > 300 ? t('<div class="small faint" style="padding:8px 14px">\u2026\u5171 ') + rows.length + t(" \u884C\u2014\u2014\u7528\u8FC7\u6EE4\u6536\u7A84\u3002</div>") : "") + "</div>" : emptyBox(t("\u6682\u65E0\u6587\u6863"), t("\u8FC7\u6EE4\u96F6\u547D\u4E2D\u6216\u5FEB\u7167\u4E3A\u7A7A\u3002"))) + '<div class="small faint" style="margin-top:8px">' + esc(t("tag \u6765\u81EA\u4E09\u6E90\u771F\u5B9E\u805A\u5408\uFF08/v1/context/index \u222A \u5BF9\u8C61 tags \u222A \u9884\u53D6 24 \u7BC7\u7F13\u5B58 frontmatter\uFF09\uFF1B\u91CD\u8981\u5EA6\u6765\u81EA\u7F13\u5B58\u6B63\u6587\u3002")) + "</div></div>";
    const inp = el.querySelector("#docs-q");
    if (inp) inp.addEventListener("input", () => {
      docFilter.q = inp.value;
      docsScreen(el, ctx2);
    });
    const tagInp = el.querySelector("#docs-tag");
    if (tagInp) tagInp.addEventListener("input", () => {
      docFilter.tag = tagInp.value;
      docFilter.routeTag = "";
      docsScreen(el, ctx2);
    });
    const clr = el.querySelector("[data-clear-tag]");
    if (clr) clr.addEventListener("click", () => {
      docFilter.tag = "";
      docFilter.routeTag = "";
      ctx2.go("docs");
    });
    el.querySelectorAll("[data-docs-kind]").forEach((c) => {
      c.addEventListener("click", () => {
        docFilter.kind = c.dataset.docsKind;
        docsScreen(el, ctx2);
      });
    });
    el.querySelectorAll(".mini-row[data-doc-row]").forEach((row) => {
      row.style.cursor = "pointer";
      row.addEventListener("click", (ev) => {
        if (ev.target && ev.target.closest("[data-go]")) return;
        ctx2.go(row.dataset.docRow);
      });
    });
  }
  async function hydrateDesignDoc(el, ctx2, repo, path) {
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const cacheKey = repo + "|" + path;
    if (!dd.cache.has(cacheKey)) {
      const r = await ctx2.api.designDoc(repo, path);
      if (!el.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
      dd.cache.set(cacheKey, r);
    }
    const res = dd.cache.get(cacheKey);
    const box = document.querySelector("[data-dd-body]");
    if (!box) return;
    if (!res || !res.ok) {
      box.innerHTML = t('<div class="ob-empty">\u6B63\u6587\u6682\u4E0D\u53EF\u8BFB\uFF08') + esc(String(res && res.error || res && res.status || t("\u7F51\u7EDC"))) + t("\uFF09\u2014\u2014HEAD \u6743\u5A01\u96F6\u7F13\u5B58\uFF0C\u91CD\u8FDB\u5373\u73B0\u8BFB\u3002</div>");
      return;
    }
    const data = res.data && typeof res.data === "object" ? res.data : {};
    const content = data.content != null ? String(data.content) : typeof res.data === "string" ? res.data : "";
    const { body } = splitFrontmatter(content);
    box.innerHTML = '<div data-meta-header-slot style="margin-bottom:8px"></div><div data-dd-bodywrap></div>';
    paintMetaHeader(
      box.querySelector("[data-meta-header-slot]"),
      ctx2,
      { kind: "DesignDoc", scope: repo, name: path },
      content,
      { head: true, headSha: String(data.sha || ""), sidecar: data.meta && typeof data.meta === "object" ? data.meta : null }
    );
    box.querySelector("[data-dd-bodywrap]").innerHTML = '<div data-md-render="1">' + renderMarkdown(body) + "</div>" + rawDisclosure(t("\u539F\u6587\uFF08HEAD\uFF09"), content) + '<div class="small faint" style="margin-top:6px">' + (data.sha ? "HEAD " + esc(String(data.sha).slice(0, 8)) : "HEAD") + t("\uFF08git \u6743\u5A01\u96F6\u7F13\u5B58\uFF09 \xB7 \u8349\u7A3F/Apply \u901A\u8DEF\u7ECF\u7F16\u8F91\u5668\u5C4F\uFF08doc/\u2026\uFF09\u3002</div>");
  }
  function designDocScreen(el, ctx2) {
    const repo = ctx2.params && ctx2.params.repo;
    const path = ctx2.params && ctx2.params.path;
    if (repo && path) {
      dd.sel = { repo, path };
    }
    const sel = dd.sel;
    el.innerHTML = '<div class="view-inner">' + screenHead(t("\u8BBE\u8BA1\u6587\u6863"), t("\u767D\u540D\u5355\u4ED3\u6811\uFF08GET /v1/design-docs\u2014\u2014HEAD \u6743\u5A01\u96F6\u7F13\u5B58\uFF09")) + t('<div class="dd-grid"><div data-dd-tree><div class="small faint">\u4ED3\u6811\u73B0\u8BFB\u4E2D\u2026</div></div>') + '<div class="card"><div class="card-b" data-dd-body>' + (sel ? t('<div class="small faint">\u6B63\u6587\u73B0\u8BFB\u4E2D\u2026</div>') : emptyBox(t("\u672A\u9009\u6863"), t("\u5DE6\u4FA7\u9009\u62E9\u4ED3\u6587\u4EF6\uFF08\u9ED8\u8BA4\u9996\u4ED3\u9996\u6587\u4EF6\uFF09\u3002"))) + "</div></div></div></div>";
    ctx2.api.designDocs().then((r) => {
      if (!el.isConnected) return;
      const treeBox = el.querySelector("[data-dd-tree]");
      if (!treeBox) return;
      if (!r.ok) {
        treeBox.innerHTML = t('<div class="ob-empty">\u4ED3\u6811\u6682\u4E0D\u53EF\u8BFB\uFF08') + esc(String(r.error || r.status)) + "\uFF09\u3002</div>";
        return;
      }
      dd.tree = r.data;
      const files = Array.isArray(r.data) ? r.data : Array.isArray(r.data && r.data.docs) ? r.data.docs : r.data && r.data.files || [];
      if (!files.length) {
        treeBox.innerHTML = t('<div class="ob-empty">\u4ED3\u6811\u4E3A\u7A7A\u3002</div>');
        return;
      }
      treeBox.innerHTML = files.map((f) => {
        const fr = f.repo || (r.data.repo || "");
        const fp = f.path || String(f);
        const key = fr + "|" + fp;
        return '<div class="mini-row' + (sel && sel.repo === fr && sel.path === fp ? " sel" : "") + '" data-dd-file="' + esc(key) + '"><div class="mr-main"><div class="mono small">' + esc(trunc(fp, 40)) + '</div><div class="mr-sub tiny">' + esc(fr) + "</div></div></div>";
      }).join("");
      treeBox.querySelectorAll("[data-dd-file]").forEach((row) => {
        row.addEventListener("click", () => {
          const [fr, ...rest] = row.dataset.ddFile.split("|");
          const fp = rest.join("|");
          ctx2.go("designdoc/" + encodeURIComponent(fr) + "/" + encodeURIComponent(fp));
        });
      });
      if (!sel && files.length) {
        const f0 = files[0];
        ctx2.go("designdoc/" + encodeURIComponent(String(f0.repo || r.data.repo || "")) + "/" + encodeURIComponent(String(f0.path || f0)));
        return;
      }
      if (sel) hydrateDesignDoc(el, ctx2, sel.repo, sel.path);
    }).catch(() => {
    });
  }
  function deliveriesPanelHtml(data) {
    if (!data || typeof data !== "object") {
      return '<div class="small faint">' + esc(t("\u9001\u8FBE\u6295\u5F71\u9762\u672A\u5C31\u7EEA\uFF08GET /v1/notifications/deliveries \u7F3A\u5E2D\u2014\u2014\u4E0D\u5192\u5145\u7A7A\uFF09\u3002")) + "</div>";
    }
    const rows = Array.isArray(data.deliveries) ? data.deliveries : [];
    const summary = '<div style="display:flex;gap:6px;flex-wrap:wrap;margin-bottom:8px">' + chip(t("\u603B ") + String(data.count != null ? data.count : rows.length), "mute") + (data.accepted != null ? chip("accepted " + data.accepted, "warn") : "") + (data.sent != null ? chip("sent " + data.sent, "ok") : "") + (data.failed != null ? chip("failed " + data.failed, Number(data.failed) > 0 ? "esc" : "mute") : "") + "</div>";
    if (!rows.length) {
      return summary + '<div class="small faint">' + esc(t("\u9001\u8FBE\u6295\u5F71\u6682\u65E0\u8BB0\u5F55\uFF08accepted\u2192sent/failed \u51FA\u7AD9\u65F6\u6B64\u5904\u5B9E\u65F6\u843D\u884C\uFF09\u3002")) + "</div>";
    }
    const stCls = (st) => st === "sent" ? "ok" : st === "failed" ? "esc" : st === "accepted" ? "warn" : "mute";
    return summary + rows.slice(0, 30).map((d) => {
      const st = String(d && d.status || "unknown");
      const att = d && d.attempts != null ? Number(d.attempts) : 0;
      const eid = d && d.eventId ? String(d.eventId) : "";
      const goAttr = eid ? ' data-go="event/' + encodeURIComponent(eid) + '" role="link" style="cursor:pointer"' : "";
      return '<div class="mini-row"' + goAttr + '><div class="mr-main"><div>' + chip(st, stCls(st)) + " <b>" + esc(String(d && d.type || "\u2014")) + "</b> " + chip("channel: " + String(d && d.channel || "*"), "mute") + " " + chip("attempts: " + att, att > 0 ? "acc" : "mute") + '</div><div class="mr-sub small mono">' + esc(trunc(String(d && (d.dedupKey || d.route) || "\u2014"), 64)) + (d && d.lastError ? " \xB7 " + esc(trunc(String(d.lastError), 60)) : "") + '</div></div><span class="small faint">' + esc(relAge(d && (d.completedAt || d.at)) || "") + "</span></div>";
    }).join("");
  }
  function connectorsScreen(el, ctx2) {
    const kinds = ["ConnectorOps", "WorkService", "Provider", "Ref", "Notification"];
    const secs = kinds.map((kind) => {
      const rows = ctx2.store.refs(kind);
      return '<section class="card"><div class="card-h"><span class="h-title">' + esc(kind) + '</span><span class="h-right">' + chip(rows.length + t(" \u4E2A"), rows.length ? "acc" : "mute") + '</span></div><div class="card-b">' + (rows.length ? rows.map((o) => '<div class="mini-row" data-conn-row="' + esc(docRoute(kind, o.scope, o.name)) + '"><div class="mr-main"><div><b class="mono">' + esc(o.name) + '</b></div><div class="mr-sub small mono">' + esc(o.scope) + " \xB7 r" + esc(String(o.rv || 0)) + '</div></div><span class="ws-chip">' + chipGo(t("\u5DE5\u4F5C\u53F0"), objRoute(kind, o.scope, o.name), "mute", true) + "</span></div>").join("") : t('<div class="small faint">\u8BE5 kind \u6682\u65E0\u5728\u518C\u5BF9\u8C61\u3002</div>')) + "</div></section>";
    }).join("");
    const snapDel = ctx2.store.deliveries ? ctx2.store.deliveries() : null;
    el.innerHTML = '<div class="view-inner">' + screenHead(
      t("\u8FDE\u63A5\u5C42"),
      t("ConnectorOps / WorkService / Provider / Ref / Notification \u4E94\u65CF + \u901A\u77E5\u9001\u8FBE\u6295\u5F71 + \u7ED1\u5B9A\u5DE5\u4F5C\u6D41\u4E92\u94FE"),
      chipGo(t("\u5DE5\u4F5C\u6D41"), "workflows", "mute", false)
    ) + '<div class="grid cols-2">' + secs + '</div><h2 class="sec-title" style="margin-top:16px">' + esc(t("\u901A\u77E5\u9001\u8FBE\u6295\u5F71\uFF08/v1/notifications/deliveries\uFF09")) + '</h2><div class="card"><div class="card-b" data-conn-deliveries>' + (snapDel ? deliveriesPanelHtml(snapDel) : '<div class="small faint">' + esc(t("\u9001\u8FBE\u6295\u5F71\u73B0\u8BFB\u4E2D\u2026")) + "</div>") + '</div></div><h2 class="sec-title" style="margin-top:16px">' + esc(t("\u7ED1\u5B9A\u5DE5\u4F5C\u6D41\uFF08workflows \u2194 connectors \u4E92\u94FE\uFF09")) + '</h2><div class="card"><div class="card-b" data-conn-wfs><div class="small faint">' + esc(t("\u5DE5\u4F5C\u6D41\u7ED1\u5B9A\u73B0\u8BFB\u4E2D\u2026")) + "</div></div></div></div>";
    el.querySelectorAll(".mini-row[data-conn-row]").forEach((row) => {
      row.style.cursor = "pointer";
      row.addEventListener("click", (ev) => {
        if (ev.target && ev.target.closest("[data-go]")) return;
        ctx2.go(row.dataset.connRow);
      });
    });
    const stamp = typeof location !== "undefined" ? location.hash : "";
    if (ctx2.api && typeof ctx2.api.deliveries === "function") {
      ctx2.api.deliveries().then((r) => {
        if (!el.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
        const box = el.querySelector("[data-conn-deliveries]");
        if (!box) return;
        if (r && r.ok) box.innerHTML = deliveriesPanelHtml(r.data);
        else if (!snapDel) box.innerHTML = '<div class="small faint">' + esc(t("\u9001\u8FBE\u6295\u5F71\u6682\u4E0D\u53EF\u8BFB\uFF08")) + esc(String(r && (r.error || r.status) || "")) + "\uFF09</div>";
      }).catch(() => {
      });
    }
    if (ctx2.api && typeof ctx2.api.workflows === "function") {
      ctx2.api.workflows().then((r) => {
        if (!el.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
        const box = el.querySelector("[data-conn-wfs]");
        if (!box) return;
        const wfs = r && r.ok && r.data && Array.isArray(r.data.workflows) ? r.data.workflows : [];
        if (!wfs.length) {
          box.innerHTML = '<div class="small faint">' + esc(t("\u6682\u65E0\u5728\u518C\u5DE5\u4F5C\u6D41\u7ED1\u5B9A\u3002")) + "</div>";
          return;
        }
        box.innerHTML = wfs.map((w) => {
          const nm = String(w && w.name || "");
          const ex = w && w.execution || {};
          const conn = String(ex.connector || "") + (ex.connectorScope ? "/" + String(ex.connectorScope) : "");
          return '<div class="mini-row" data-go="wf/' + encodeURIComponent(nm) + '" style="cursor:pointer"><div class="mr-main"><div><b class="mono">' + esc(nm) + "</b> " + chip(conn || t("\u672A\u7ED1\u5B9A"), conn ? "acc" : "mute") + '</div><div class="mr-sub small">' + esc(String(ex.state || "\u2014")) + '</div></div><span class="ws-chip">' + chipGo(t("\u5DE5\u4F5C\u6D41\u8BE6\u60C5"), "wf/" + encodeURIComponent(nm), "mute", true) + "</span></div>";
        }).join("");
      }).catch(() => {
      });
    }
  }
  function proposalsScreen(el, ctx2) {
    const rows = ctx2.store.proposals();
    const focusName = String(ctx2.params && ctx2.params.name || "").trim();
    if (focusName) propOpen.add(focusName);
    const cards = rows.map((p, i) => {
      const keys = p && typeof p === "object" ? Object.keys(p) : [];
      const key = String(p && p.name || "") || "#" + i;
      const open = propOpen.has(key);
      const isFocus = !!focusName && key === focusName;
      const known = ["title", "summary", "status", "proposer", "createdAt", "scope", "name"];
      const sub = known.filter((k) => p[k] != null && k !== "title").map((k) => esc(k) + ": " + esc(trunc(String(p[k]), 40))).join(" \xB7 ");
      const extras = keys.filter((k) => !known.includes(k) && k !== "summary" && p[k] != null && p[k] !== "").map((k) => '<span class="k">' + esc(k) + '</span><span class="v">' + esc(trunc(String(p[k]), 200)) + "</span>").join("");
      const approveRoute = p && p.name ? "approvals/" + encodeURIComponent(String(p.name)) : "approvals";
      const propLink = p && p.name ? "prop/" + encodeURIComponent(String(p.name)) : "";
      return '<div class="mini-row" data-prop="' + esc(key) + '"' + (isFocus ? ' data-prop-focused="1" style="outline:2px solid var(--accent)"' : "") + '><div class="mr-main"><div><b>' + esc(String(p.title || p.name || t("\u672A\u547D\u540D\u63D0\u6848"))) + "</b> " + chip(String(p.status || t("\u5F85\u5B9A")), "mute") + '</div><div class="mr-sub small">' + (sub || t('<span class="faint">\u8981\u7D20\u5B57\u6BB5\u7F3A\u5E2D</span>')) + "</div>" + (open ? '<div class="mr-sub" style="margin-top:6px;white-space:pre-wrap">' + esc(trunc(String(p.summary || ""), 2e3) || t("summary \u7F3A\u5E2D\u2014\u2014\u8BE6\u60C5\u89C1\u5BF9\u8C61\u5DE5\u4F5C\u53F0")) + "</div>" + (extras ? '<div class="kv" style="margin-top:6px">' + extras + "</div>" : "") + '<div style="margin-top:6px;display:flex;gap:6px;flex-wrap:wrap">' + (p && p.name ? chipGo(t("\u5BF9\u8C61\u5DE5\u4F5C\u53F0"), objRoute("Proposal", p.scope || "global", p.name), "mute", true) : "") + (propLink ? chipGo(t("\u76F4\u8FBE\u94FE\u63A5"), propLink, "mute", true) : "") + '</div><div class="small faint" style="margin-top:4px">' + esc(t("\u518D\u70B9\u4E00\u6B21\u5361\u7247\u6536\u8D77\u3002")) + "</div>" : "") + '</div><span class="ws-chip">' + chipGo(t("\u53BB\u5BA1\u6279"), approveRoute, "acc", true, t("\u5B9A\u4F4D\u81F3\u5BA1\u6279\u9875\u8BE5\u63D0\u6848")) + "</span></div>";
    }).join("");
    el.innerHTML = '<div class="view-inner">' + screenHead(
      t("\u9700\u6C42\u63D0\u6848"),
      t("\u63D0\u6848\u9762\uFF08\u70B9\u5361\u7247\u5C55\u5F00 summary \u5168\u6587\u4E0E\u8981\u7D20\uFF1B\u652F\u6301 #/prop/:name \u76F4\u8FBE\uFF09\u2014\u2014\u6279\u51C6/\u9A73\u56DE\u8D70\u5BA1\u6279\u4E0E Apply\uFF08LLM \u65E0 Apply \u6743\uFF09"),
      focusName ? chipGo(t("\u5168\u90E8\u63D0\u6848"), "proposals", "mute", false) : ""
    ) + (rows.length ? '<div class="card">' + cards + "</div>" : emptyBox(t("\u6682\u65E0\u63D0\u6848"), t("\u5FEB\u7167 proposals \u805A\u5408\u4E3A\u7A7A\u3002"))) + "</div>";
    el.querySelectorAll("[data-prop]").forEach((row) => {
      row.addEventListener("click", (ev) => {
        if (ev.target.closest("[data-go]")) return;
        const key = row.dataset.prop;
        if (propOpen.has(key)) propOpen.delete(key);
        else propOpen.add(key);
        proposalsScreen(el, ctx2);
      });
    });
    if (focusName) {
      const hit = el.querySelector('[data-prop-focused="1"]');
      if (hit && hit.scrollIntoView) hit.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  }
  function systemExtraBlocks(ctx2, scope) {
    const tasks = ctx2.store.tasks().filter((t2) => t2 && (String(t2.scope || "") === scope || String(t2.spec && t2.spec.scopeRef && t2.spec.scopeRef.name || "") === scope));
    const drafts = ctx2.store.drafts().filter((d) => d && String(d.scope || "global") === scope);
    const runs = ctx2.store.runs().filter((r) => {
      if (!r) return false;
      let hay = "";
      try {
        hay = JSON.stringify([r.scope, r.ownerScope, r.involved, r.route, r.workflow]);
      } catch {
        hay = "";
      }
      return hay.includes(scope);
    });
    const sec = (title, bodyHtml, emptyText) => '<h2 class="sec-title">' + esc(title) + '</h2><div class="card">' + (bodyHtml || '<div class="card-b"><div class="small faint">' + esc(emptyText) + "</div></div>") + "</div>";
    const taskHtml = tasks.slice(0, 8).map((t2) => {
      const id = (t2.scope || "global") + "/" + (t2.name || "");
      const ph = String(t2.status && t2.status.phase || "");
      return '<div class="mini-row" data-go="task/' + encodeURIComponent(id) + '" style="cursor:pointer"><div class="mr-main"><div><b>' + esc(String(t2.spec && t2.spec.title || t2.name || id)) + "</b> " + (ph ? chip(ph, ph === "running" || ph === "verifying" ? "acc" : "mute") : "") + '</div><div class="mr-sub small mono">' + esc(id) + "</div></div></div>";
    }).join("");
    const draftHtml = drafts.slice(0, 8).map((d) => {
      const sc = d.scope || "global";
      return '<div class="mini-row" data-go="' + esc(docRoute(d.kind, sc, d.name)) + '" style="cursor:pointer"><div class="mr-main"><div><b>' + esc(String(d.name || "")) + "</b> " + kindBadge(String(d.kind || "")) + " " + chip(t("\u8349\u7A3F"), "warn") + '</div><div class="mr-sub small mono">' + esc(String(d.fieldManager || "\u2014")) + '</div></div><span class="ws-chip">' + chipGo(t("\u53BB\u5BA1\u6279"), "approvals/" + encodeURIComponent(d.kind + "/" + sc + "/" + d.name), "acc", true) + "</span></div>";
    }).join("");
    const runHtml = runs.slice(0, 8).map((r) => {
      const wf = String(r.workflow || r.name || "");
      const no = r.run != null ? r.run : r.no != null ? r.no : "";
      const go = wf && no !== "" && no != null ? "wf/" + encodeURIComponent(wf) + "/r/" + encodeURIComponent(String(no)) : wf ? "wf/" + encodeURIComponent(wf) : "";
      return '<div class="mini-row"' + (go ? ' data-go="' + esc(go) + '" style="cursor:pointer"' : "") + '><div class="mr-main"><div><b class="mono">' + esc(wf || "\u2014") + "</b> " + (no !== "" ? '<span class="small faint">#' + esc(String(no)) + "</span>" : "") + '</div><div class="mr-sub small mono">' + esc(trunc(String(r.involved || r.route || "\u2014"), 56)) + "</div></div></div>";
    }).join("");
    return sec(t("\u76F8\u5173\u4EFB\u52A1"), taskHtml, t("\u6682\u65E0\u76F8\u5173\u4EFB\u52A1\u3002")) + sec(t("\u8349\u7A3F"), draftHtml, t("\u6682\u65E0\u8349\u7A3F\u3002")) + sec(t("\u8FD0\u884C\u8BB0\u5F55"), runHtml, t("\u6682\u65E0\u8FD0\u884C\u8BB0\u5F55\u3002"));
  }
  function systemScreen(el, ctx2) {
    const scope = String(ctx2.params.name || "");
    const objs = ctx2.store.objects().filter((o) => o.scope === scope);
    const evs = ctx2.store.events().filter((e) => String(e.involved || "").includes(scope)).slice(0, 10);
    const byKind = {};
    objs.forEach((o) => {
      byKind[o.kind] = (byKind[o.kind] || 0) + 1;
    });
    el.innerHTML = '<div class="view-inner">' + screenHead(scope, "scope=" + scope + t(" \u7684\u5BF9\u8C61\u3001\u4EFB\u52A1\u3001\u8349\u7A3F\u3001\u8FD0\u884C\u4E0E\u4E8B\u4EF6\uFF08\u5FEB\u7167\u805A\u5408\u6295\u5F71\uFF09"), chipGo(t("\u7CFB\u7EDF\u6811"), "tree", "mute", false)) + '<div class="statbar">' + Object.keys(byKind).sort().map((k) => '<div class="stat" data-go="' + esc("obj/" + encodeURIComponent(k)) + '" role="link"><div class="st-num">' + byKind[k] + '</div><div class="st-label">' + esc(k) + "</div></div>").join("") + "</div>" + (objs.length ? t('<h2 class="sec-title">\u5BF9\u8C61</h2><div class="card">') + objs.map((o) => '<div class="mini-row" data-go="' + esc(objRoute(o.kind, o.scope, o.name)) + '"><div class="mr-main"><div><b>' + esc(o.name) + "</b> " + kindBadge(o.kind) + '</div><div class="mr-sub small mono">r' + esc(String(o.rv || 0)) + "</div></div></div>").join("") + "</div>" : emptyBox(t("\u8BE5 scope \u6682\u65E0\u5728\u518C\u5BF9\u8C61"), t("\u8BE5 scope \u4E0B\u6682\u65E0\u5BF9\u8C61\u3002"))) + (evs.length ? t('<h2 class="sec-title">\u6700\u8FD1\u76F8\u5173\u4E8B\u4EF6</h2><div class="card">') + evs.map((e) => '<div class="mini-row" data-go="event/' + encodeURIComponent(String(e.id ?? "")) + '"><div class="mr-main"><div>' + esc(String(e.type || "")) + '</div><div class="mr-sub small mono">' + esc(trunc(String(e.involved || "\u2014"), 60)) + "</div></div></div>").join("") + "</div>" : "") + systemExtraBlocks(ctx2, scope) + "</div>";
  }
  function render11(el, ctx2) {
    const r = ctx2.route;
    if (r === "tree") treeScreen(el, ctx2);
    else if (r === "docs") docsScreen(el, ctx2);
    else if (r === "designdoc") designDocScreen(el, ctx2);
    else if (r === "connectors") connectorsScreen(el, ctx2);
    else if (r === "proposals" || r === "prop") proposalsScreen(el, ctx2);
    else if (r === "s") systemScreen(el, ctx2);
    else el.innerHTML = '<div class="view-inner">' + emptyBox(t("\u672A\u77E5\u8584\u5C4F"), String(r)) + "</div>";
  }
  function renderBiz(el, ctx2) {
    void ctx2;
    el.innerHTML = '<div class="view-inner"><div class="biz-hero"><div class="biz-hero-mark">B</div><h2 class="sec-title">' + esc(t("\u4E1A\u52A1\u63A7\u5236\u53F0")) + '</h2><div class="small faint">' + esc(t("\u4E1A\u52A1\u5C42\u672A\u63A5\u5165\u2014\u2014\u540E\u7EED\u4E1A\u52A1\u670D\u52A1\u767B\u5F55\u540E\u5728\u6B64\u627F\u8F7D\u4E1A\u52A1\u5927\u5C4F\u4E0E\u64CD\u4F5C\u9762\u3002")) + '</div><div style="margin-top:10px">' + chipGo(t("\u67E5\u770B\u4E1A\u52A1\u8BBE\u8BA1\u7A3F"), "designdoc/" + encodeURIComponent("core_design") + "/" + encodeURIComponent("concepts/business-system.md"), "mute", true) + "</div></div></div>";
  }

  // src/screens/audit.js
  function auditCardHtml(a) {
    const findings = Array.isArray(a && a.findings) ? a.findings : [];
    const unbound = Array.isArray(a && a.unbound) ? a.unbound : [];
    const skipped = Array.isArray(a && a.skipped) ? a.skipped : [];
    const errN = findings.filter((f) => f && f.severity === "error").length;
    const warnN = findings.length - errN;
    const ok = !!(a && a.ok);
    const tool = a && a.tool && a.tool.name ? String(a.tool.name) + (a.tool.version ? "@" + String(a.tool.version) : "") : "connectops-code-audit";
    const ranAt = String(a && a.ranAt || "");
    const findHtml = findings.length ? '<div style="overflow-x:auto;margin-top:8px"><table class="tbl"><thead><tr><th>' + esc(t("\u89C4\u5219")) + "</th><th>" + esc(t("\u7EA7\u522B")) + "</th><th>" + esc(t("\u4F4D\u7F6E")) + "</th><th>" + esc(t("\u8BF4\u660E")) + "</th><th>checker</th></tr></thead><tbody>" + findings.slice(0, 100).map((f) => {
      const sev = String(f && f.severity || "error");
      const loc = String(f && f.file || "\u2014") + (f && Number(f.line) > 0 ? ":" + String(f.line) : "");
      return '<tr><td class="mono small">' + esc(String(f && f.rule || "\u2014")) + "</td><td>" + chip(sev, sev === "error" ? "esc" : "warn") + '</td><td class="mono small">' + esc(trunc(loc, 56)) + '</td><td class="small">' + esc(trunc(String(f && f.message || "\u2014"), 120)) + '</td><td class="mono small faint">' + esc(String(f && f.checkRef || "\u2014")) + "</td></tr>";
    }).join("") + "</tbody></table></div>" : '<div class="small faint" style="margin-top:6px">' + esc(t("findings \u4E3A\u7A7A\uFF08\u96F6\u4EE3\u7801\u7EA7\u8FDD\u89C4\u53D1\u73B0\uFF09\u3002")) + "</div>";
    const unboundHtml = unbound.length ? '<div style="margin-top:8px"><div class="small" style="font-weight:600;margin-bottom:4px">' + esc(t("\u672A\u7ED1\u5B9A\u68C0\u67E5\u5668\uFF08unbound\u2014\u2014\u663E\u5F0F\u963B\u65AD\uFF0C\u4E0D\u8DF3\u8FC7\uFF09\uFF1A")) + "</div>" + unbound.map((u) => '<div class="mini-row"><div class="mr-main"><div>' + chip("unbound", "esc") + ' <b class="mono">' + esc(String(u && u.rule || "*")) + "</b> " + (u && u.checkRef ? chip(String(u.checkRef), "mute") : "") + '</div><div class="mr-sub small">' + esc(String(u && u.reason || "\u2014")) + "</div></div></div>").join("") + "</div>" : "";
    const skippedHtml = skipped.length ? '<details class="ex-block" style="margin-top:8px"><summary class="ex-sum"><span class="ex-chev">\u25B8</span>' + esc(t("\u975E code \u89C4\u5219\u5BF9\u8D26\uFF08agent/human \u8DEF\u7531 \xB7 ")) + skipped.length + '\uFF09</summary><div class="ex-body">' + skipped.map((s) => '<div class="small mono">' + chip(String(s && s.mode || "\u2014"), "mute") + " " + esc(String(s && s.rule || "\u2014")) + ' \u2014 <span class="faint">' + esc(String(s && s.reason || "")) + "</span></div>").join("") + "</div></details>" : "";
    return '<section class="card" style="margin-bottom:12px"><div class="card-h"><span class="h-title mono">' + esc(String(a && a.auditId || "\u2014")) + '</span><span class="h-right">' + chip(ok ? t("\u901A\u8FC7") : t("\u963B\u65AD"), ok ? "ok" : "esc") + " " + chip("error " + errN, errN ? "esc" : "mute") + " " + chip("warn " + warnN, warnN ? "warn" : "mute") + (unbound.length ? " " + chip("unbound " + unbound.length, "esc") : "") + '</span></div><div class="card-b"><div class="small faint">' + esc(tool) + (ranAt ? " \xB7 " + esc(ranAt.replace("T", " ").slice(0, 19)) + (relAge(ranAt) ? "\uFF08" + esc(relAge(ranAt)) + "\uFF09" : "") : "") + "</div>" + findHtml + unboundHtml + skippedHtml + "</div></section>";
  }
  function renderAuditPayload(data) {
    if (!data || typeof data !== "object") {
      return '<div class="card"><div class="card-b">' + emptyBox(t("\u4EE3\u7801\u5BA1\u6838\u53EA\u8BFB\u9762\u672A\u88C5\u914D"), t("\u540E\u7AEF\u672A\u66B4\u9732 /v1/code-audit/findings\uFF08\u65E7\u7248 cored \u6216\u9A71\u52A8\u672A\u88C5\u914D\u2014\u2014\u8BDA\u5B9E\u964D\u7EA7\uFF0C\u4E0D\u5192\u5145\u901A\u8FC7\uFF09\u3002")) + "</div></div>";
    }
    const audits = Array.isArray(data.audits) ? data.audits : [];
    if (!audits.length) {
      return '<div class="card"><div class="card-b">' + emptyBox(t("\u6682\u65E0\u4EE3\u7801\u5BA1\u6838\u8BB0\u5F55"), t("connectops-code-audit \u8FDB\u7A0B\u5185 LRU \u4E3A\u7A7A\u2014\u2014Verify \u76F8\u4F4D\u6216\u5BA1\u6838 op \u6267\u884C\u540E\u6B64\u5904\u5B9E\u65F6\u5C55\u793A findings/unbound/skipped\uFF08\u96F6\u5047\u6570\u636E\uFF09\u3002")) + "</div></div>";
    }
    return '<div class="small faint" style="margin-bottom:8px">' + esc(t("\u6570\u636E\u6E90\uFF1A")) + esc(String(data.source || "connectops-code-audit")) + esc(t(" \u53EA\u8BFB\u6295\u5F71\uFF08\u6700\u8FD1 ")) + audits.length + esc(t(" \u6B21\u5BA1\u6838\uFF1B\u53EA\u8BFB\u4E0D\u6539\u4EE3\u7801\u3001\u4E0D\u66FF\u4EE3\u5BA1\u6279\uFF09\u3002")) + "</div>" + audits.map(auditCardHtml).join("");
  }
  function render12(el, ctx2) {
    const snap = ctx2.store && ctx2.store.codeAudit ? ctx2.store.codeAudit() : null;
    el.innerHTML = '<div class="view-inner">' + screenHead(
      t("\u4EE3\u7801\u5BA1\u6838"),
      t("connectops-code-audit \u53EA\u8BFB findings \u6295\u5F71\uFF08R-TYPE-3\uFF1Acode \u89C4\u5219\u673A\u68B0\u5BA1\u6838\uFF1Berror/unbound \u963B\u65AD\uFF0Cwarn \u8FDB\u62A5\u544A\uFF09"),
      chipGo(t("\u5DE5\u4F5C\u6D41"), "workflows", "mute", false) + " " + chipGo(t("\u89C4\u5219\u8BBE\u8BA1"), "designdoc/" + encodeURIComponent("core_design") + "/" + encodeURIComponent("rule/rule_type.md"), "mute", false)
    ) + "<div data-audit-body>" + (snap ? renderAuditPayload(snap) : '<div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u4EE3\u7801\u5BA1\u6838 findings \u73B0\u8BFB\u4E2D\u2026\uFF08GET /v1/code-audit/findings\uFF09")) + "</div></div></div>") + "</div></div>";
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const fetchFn = ctx2.api && typeof ctx2.api.codeAuditFindings === "function" ? ctx2.api.codeAuditFindings() : Promise.resolve({ ok: false, status: 404, error: "\u63A5\u53E3\u7F3A\u5E2D" });
    fetchFn.then((r) => {
      if (!el.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
      const box = el.querySelector("[data-audit-body]");
      if (!box) return;
      if (!r.ok) {
        box.innerHTML = '<div class="card"><div class="card-b">' + emptyBox(
          t("\u4EE3\u7801\u5BA1\u6838\u53EA\u8BFB\u9762\u6682\u4E0D\u53EF\u8FBE\uFF08HTTP ") + String(r.status || "?") + "\uFF09",
          String(r.error || t("\u9700\u8981\u5E26 /v1/code-audit/findings \u53EA\u8BFB\u6295\u5F71\u7684 cored \u7248\u672C\u2014\u2014\u4E0D\u5192\u5145\u96F6\u53D1\u73B0\u3002"))
        ) + "</div></div>";
        return;
      }
      box.innerHTML = renderAuditPayload(r.data);
    }).catch(() => {
      const box = el.querySelector("[data-audit-body]");
      if (box && !snap) box.innerHTML = '<div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u4EE3\u7801\u5BA1\u6838 findings \u8BFB\u53D6\u5F02\u5E38\u2014\u2014\u4E0D\u5192\u5145\u901A\u8FC7\u3002")) + "</div></div></div>";
    });
  }

  // src/main.js
  var BASES = typeof window !== "undefined" && window.__CORE_CONSOLE_BASES__ || {};
  var store = createStore();
  var api = createApi(BASES.api || "", { apply: BASES.apply, drafts: BASES.drafts });
  var router = createRouter({ defaultRoute: "home" });
  var ctx = { store, api, router, go: (p) => router.go(p), params: {}, route: "" };
  if (typeof window !== "undefined") window.__CC_CONSOLE_CTX__ = ctx;
  var NAV_TREE = [
    {
      level: "sys",
      label: "Core Console",
      icon: "\u25A3",
      items: [
        ["home", "\u5F00\u59CB", "\u2727"],
        ["welcome", "\u603B\u89C8", "\u2302"],
        ["tasks", "\u5DE5\u4F5C\u4E0E\u6D41\u7A0B", "\u229E"],
        ["workflows", "\u6D41\u7A0B\u8FD0\u884C", "\u27F3"],
        ["approvals", "\u5BA1\u6279\u4E2D\u5FC3", "\u2713"],
        ["docs", "\u6587\u6863\u4E0E\u8D44\u4EA7", "\u25A4"]
      ],
      sub: { label: "\u7CFB\u7EDF\u5E95\u5C42", items: [
        ["chat/core-headless", "\u4F1A\u8BDD", "\u{1F4AC}"],
        ["tree", "\u7CFB\u7EDF\u6811", "\u2443"],
        ["tags", "\u6807\u7B7E", "#"],
        ["designdoc", "\u8BBE\u8BA1\u6587\u6863", "\u270E"],
        ["obj", "\u5BF9\u8C61\u5DE5\u4F5C\u53F0", "\u25AD"],
        ["connectors", "\u8FDE\u63A5", "\u2301"],
        ["events", "\u4E8B\u4EF6\u6D41", "\u26A1"],
        ["proposals", "\u63D0\u6848", "\u2726"],
        ["report", "\u5173\u6CE8\u62A5\u544A", "\u2248"],
        ["audit", "\u4EE3\u7801\u5BA1\u6838", "\u2713"]
      ] }
    }
  ];
  var curLevel = "sys";
  try {
    curLevel = typeof localStorage !== "undefined" && localStorage.getItem("cc.level") || "sys";
  } catch {
  }
  function langToggleHtml() {
    const cur = getLocale();
    const btn = (id, label) => '<button class="btn btn-ghost sm' + (cur === id ? " active" : "") + '" data-lang="' + id + '">' + label + "</button>";
    return '<span class="lang-toggle">' + btn("zh", "\u4E2D\u6587") + btn("en", "EN") + "</span>";
  }
  var foldState = { "\u7CFB\u7EDF\u5E95\u5C42": false };
  function approvalBadge() {
    try {
      const n = (store.drafts() || []).length + (store.events() || []).filter((e) => e && e.type === "ApprovalNeeded").length;
      return n > 0 ? '<span class="nav-badge">' + n + "</span>" : "";
    } catch {
      return "";
    }
  }
  function navTree() {
    return NAV_TREE;
  }
  function levelTabsHtml() {
    return navTree().map((sec) => {
      const on = curLevel === sec.level;
      return '<button class="level-tab' + (on ? " active" : "") + '" data-level="' + esc(sec.level) + '"><span class="nav-ico">' + sec.icon + "</span>" + esc(t(sec.label)) + "</button>";
    }).join("");
  }
  function sidebarHtml() {
    const sec = navTree().find((s) => s.level === curLevel) || navTree()[0];
    if (!sec) return "";
    const items = sec.items.map(([r, label, icon]) => {
      const badge = r === "approvals" ? approvalBadge() : "";
      return '<div class="nav-item l2" data-nav="' + esc(r.split("/")[0]) + '" data-go="' + esc(r) + '"><span class="nav-ico">' + icon + "</span>" + esc(t(label)) + badge + "</div>";
    }).join("");
    const sub = sec.sub ? '<div class="nav-sub-h" data-fold="' + esc(sec.sub.label) + '"><span class="nav-ico fold-chev">' + (foldState[sec.sub.label] ? "\u25B8" : "\u25BE") + "</span>" + esc(t(sec.sub.label)) + "</div>" + (foldState[sec.sub.label] ? "" : '<nav class="side-nav l3">' + sec.sub.items.map(([r, label, icon]) => '<div class="nav-item l3" data-nav="' + esc(r.split("/")[0]) + '" data-go="' + esc(r) + '"><span class="nav-ico">' + icon + "</span>" + esc(t(label)) + "</div>").join("") + "</nav>") : "";
    return '<nav class="side-nav l2">' + items + "</nav>" + sub;
  }
  function sideFootHtml() {
    return '<div class="side-foot"><span class="small faint">Core Console</span></div>';
  }
  function crumbLabel(route) {
    const pattern = String(route || "");
    if (pattern.startsWith("wf/") && pattern.includes("/r/")) return t("\u8FD0\u884C\u8BE6\u60C5");
    const bare = pattern.split("/")[0];
    for (const sec of NAV_TREE) {
      for (const [r, label] of sec.items) if (r.split("/")[0] === bare) return t(label);
      if (sec.sub) {
        for (const [r, label] of sec.sub.items) if (r.split("/")[0] === bare) return t(label);
      }
    }
    const extra = { task: t("\u4EFB\u52A1\u8BE6\u60C5"), event: t("\u4E8B\u4EF6\u8BE6\u60C5"), doc: t("\u6587\u6863\u7F16\u8F91\u5668"), designdoc: t("\u8BBE\u8BA1\u6587\u6863"), wf: t("\u5DE5\u4F5C\u6D41"), session: t("\u4F1A\u8BDD"), service: t("Service \u5927\u5C4F"), s: t("\u7CFB\u7EDF\u8BE6\u60C5"), inbox: t("\u6536\u4EF6\u7BB1"), prop: t("\u9700\u6C42\u63D0\u6848"), tags: t("\u6807\u7B7E"), audit: t("\u4EE3\u7801\u5BA1\u6838") };
    return extra[bare] || bare;
  }
  function shellHtml() {
    return '<div class="app"><aside class="sidebar"><div class="side-logo"><span class="logo-mark">C</span><span><span class="logo-name">Core Console</span><br><span class="logo-sub">' + esc(t("\u7CFB\u7EDF\u5DE5\u4F5C\u53F0")) + "</span></span></div><div data-nav-tree>" + sidebarHtml() + "</div>" + sideFootHtml() + '</aside><div class="main-wrap"><header class="topbar"><nav class="level-tabs" data-level-tabs>' + levelTabsHtml() + '</nav><span class="topbar-crumb" data-topbar-crumb></span><span class="topbar-right"><span class="lamp off" data-conn-lamp-top></span><span class="small faint" data-conn-label-top>' + esc(t("\u8FDE\u63A5\u4E2D\u2026")) + "</span>" + langToggleHtml() + '</span></header><main class="main"><div class="view" id="view"></div></main></div></div>';
  }
  function paintNav(route) {
    document.querySelectorAll("[data-nav]").forEach((n) => {
      const r = String(n.dataset.nav || "");
      const on = route === r || r === "tasks" && route === "task" || r === "events" && route === "event" || r === "workflows" && route === "wf" || r === "proposals" && route === "prop" || r === "docs" && route === "doc" || r === "welcome" && route === "service" || r === "chat" && (route === "session" || route.startsWith("chat"));
      n.classList.toggle("active", on);
    });
    const wantLevel = (function() {
      const hit = navTree().find((s) => s.items.some(([r]) => r === route || r.split("/")[0] === route));
      return hit ? hit.level : null;
    })();
    if (wantLevel && wantLevel !== curLevel) {
      curLevel = wantLevel;
      try {
        localStorage.setItem("cc.level", curLevel);
      } catch {
      }
      const tree = document.querySelector("[data-nav-tree]");
      if (tree) tree.innerHTML = sidebarHtml();
      const tabs = document.querySelector("[data-level-tabs]");
      if (tabs) tabs.innerHTML = levelTabsHtml();
    }
  }
  function screenOf(fn) {
    return (el, params) => {
      ctx.params = params;
      ctx.route = router.current.route;
      try {
        fn(el, ctx);
      } catch (err) {
        try {
          const msg = String(err && err.message || err || "unknown");
          el.innerHTML = '<div class="view-inner"><div class="dh"><div class="dh-main"><div class="dh-title"><h1>Screen Error</h1></div><div class="dh-sub">\u8BE5\u5C4F\u6E32\u67D3\u5931\u8D25\u2014\u2014\u9519\u8BEF\u5DF2\u663E\u5F0F\u5448\u73B0\uFF08\u4E0D\u9759\u9ED8\u5361\u52A0\u8F7D\uFF1B\u8FD4\u56DE\u6216\u6362\u5C4F\u5373\u6062\u590D\uFF09\u3002</div></div></div><div class="card"><div class="card-b"><div class="small faint mono" style="white-space:pre-wrap">' + String(msg).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[c]) + '</div><div style="margin-top:10px"><button class="btn btn-secondary sm" data-go="welcome">\u56DE\u603B\u89C8</button> <button class="btn btn-ghost sm" data-retry-screen="1">\u91CD\u8BD5\u672C\u5C4F</button></div></div></div></div>';
          const retry = el.querySelector("[data-retry-screen]");
          if (retry) retry.addEventListener("click", () => {
            try {
              fn(el, ctx);
            } catch {
            }
          });
        } catch {
        }
      }
      paintNav(router.current.route);
      try {
        const crumb = document.querySelector("[data-topbar-crumb]");
        if (crumb) crumb.textContent = crumbLabel(router.current.pattern || router.current.route);
      } catch {
      }
    };
  }
  var lastConnState = "connecting";
  function paintConn(state) {
    lastConnState = state;
    const text2 = { connected: t("\u5B9E\u65F6\uFF08WS patch\uFF09"), reconnecting: t("\u91CD\u8FDE\u4E2D\u2026"), degraded: t("\u964D\u7EA7\u8F6E\u8BE2\uFF0830s\uFF09"), connecting: t("\u8FDE\u63A5\u4E2D\u2026") }[state] || state;
    const cls = { connected: "ok", reconnecting: "warn", degraded: "warn", connecting: "off" }[state] || "off";
    [["[data-conn-lamp]", "[data-conn-label]"], ["[data-conn-lamp-top]", "[data-conn-label-top]"]].forEach(([ls, ltx]) => {
      document.querySelectorAll(ls).forEach((lamp) => {
        lamp.className = "lamp " + cls + (state === "connected" ? " pulse-ok" : "");
      });
      document.querySelectorAll(ltx).forEach((label) => {
        label.textContent = text2;
      });
    });
  }
  var hydrating = false;
  async function hydrateDocs() {
    if (hydrating) return;
    hydrating = true;
    try {
      const stale = store.staleDocs(24);
      let i = 0;
      const worker = async () => {
        while (i < stale.length) {
          const r = stale[i++];
          const res = await api.object(r.kind, r.scope, r.name);
          if (res.ok) {
            const data = res.data && typeof res.data === "object" ? res.data : {};
            const content = data.content != null ? String(data.content) : typeof res.data === "string" ? res.data : JSON.stringify(res.data, null, 2);
            store.cacheDoc(r.kind, r.scope, r.name, r.rv, content);
          }
        }
      };
      await Promise.all(Array.from({ length: Math.min(4, stale.length) }, worker));
      if (stale.length && (router.current.route === "tags" || router.current.route === "docs" || router.current.route === "obj")) {
        router.render();
      }
    } finally {
      hydrating = false;
    }
  }
  function mount(root, opts = {}) {
    if (!root) return null;
    root.classList.add("cc-app");
    root.innerHTML = shellHtml();
    const view = root.querySelector("#view");
    store.subscribe(() => {
      if (view.isConnected) router.render();
      try {
        const tree = root.querySelector("[data-nav-tree]");
        if (tree) tree.innerHTML = sidebarHtml();
      } catch {
      }
    });
    root.addEventListener("click", (ev) => {
      const lv = ev.target.closest("[data-level]");
      if (lv) {
        ev.preventDefault();
        curLevel = lv.dataset.level;
        try {
          localStorage.setItem("cc.level", curLevel);
        } catch {
        }
        const tree = root.querySelector("[data-nav-tree]");
        if (tree) tree.innerHTML = sidebarHtml();
        const tabs = root.querySelector("[data-level-tabs]");
        if (tabs) tabs.innerHTML = levelTabsHtml();
        router.go("home");
        return;
      }
      const fold = ev.target.closest("[data-fold]");
      if (fold) {
        ev.preventDefault();
        const g = fold.dataset.fold;
        foldState[g] = !foldState[g];
        const tree = root.querySelector("[data-nav-tree]");
        if (tree) tree.innerHTML = sidebarHtml();
        return;
      }
      const langBtn = ev.target.closest("[data-lang]");
      if (langBtn) {
        ev.preventDefault();
        setLocale(langBtn.dataset.lang);
        return;
      }
      const t2 = ev.target.closest("[data-go]");
      if (t2) {
        ev.preventDefault();
        router.go(t2.dataset.go);
      }
    });
    onLocaleChange(() => {
      const tree = root.querySelector("[data-nav-tree]");
      if (tree) tree.innerHTML = sidebarHtml();
      const tabs = root.querySelector("[data-level-tabs]");
      if (tabs) tabs.innerHTML = levelTabsHtml();
      paintConn(lastConnState);
      if (view.isConnected) router.render();
      try {
        document.title = "Core Console";
      } catch {
      }
    });
    router.route("home", screenOf(render)).route("welcome", screenOf(render3)).route("obj", screenOf(render7)).route("obj/:kind", screenOf(render7)).route("obj/:kind/:scope/:name", screenOf(render7)).route("events", screenOf(render6)).route("event/:id", screenOf(render6)).route("tasks", screenOf(render4)).route("task/:id", screenOf(render4)).route("report", screenOf(renderReport)).route("doc/:name", screenOf(render8)).route("doc/:kind/:scope/:name", screenOf(render8)).route("chat/:agent", screenOf(render9)).route("session/:id", screenOf(render9)).route("approvals", screenOf(render2)).route("approvals/:focus", screenOf(render2)).route("inbox", screenOf(render2)).route("proposals", screenOf(render11)).route("proposals/:name", screenOf(render11)).route("prop/:name", screenOf(render11)).route("workflows", screenOf(render5)).route("wf/:name", screenOf(render5)).route("wf/:name/r/:n", screenOf(render5)).route("tree", screenOf(render11)).route("docs", screenOf(render11)).route("docs/tag/:tag", screenOf(render11)).route("tags", screenOf(render10)).route("designdoc", screenOf(render11)).route("designdoc/:repo/:path", screenOf(render11)).route("connectors", screenOf(render11)).route("audit", screenOf(render12)).route("s/:name", screenOf(render11)).route("service/:name", screenOf(renderService)).route("biz", screenOf(renderBiz));
    router.mount(view);
    const stream = connectStream({
      url: BASES.ws || wsUrlOf(opts.location || (typeof location !== "undefined" ? location : null)),
      base: BASES.api || "",
      onSnapshot: (s) => {
        store.setConn("connected");
        store.init(s);
        hydrateDocs();
      },
      onPatch: (patch, msg) => {
        if (!store.applyPatch(patch, msg)) {
          api.snapshot().then((r) => {
            if (r.ok) {
              store.init(r.data);
              hydrateDocs();
            }
          });
        }
      },
      onState: (s) => {
        paintConn(s);
      }
    });
    return { store, api, router, stream };
  }
  if (typeof document !== "undefined") {
    const boot = () => {
      const root = document.querySelector("[data-core-console]") || document.getElementById("app");
      if (root && !root.classList.contains("cc-app")) mount(root);
    };
    if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", boot);
    else boot();
  }
  if (typeof window !== "undefined") {
    window.CoreConsoleWeb = { mount, store, api, router, connectStream, wsUrlOf };
  }
  return __toCommonJS(main_exports);
})();
