const {createApp, ref, computed, watch} = Vue
let app = createApp({
    setup() {
        const sessions = ref(getSessions())
        const selectID = getUrlParam('select',Object.keys(sessions.value)[0])
        const selected = ref()
        if (sessions.value[selectID]){
            selected.value = sessions.value[selectID]
        }else {
            selected.value = sessions.value[Object.keys(sessions.value)[0]]
        }
        const self = ref(selfInfo())
        const display = ref(false)
        return {
            sessions: sessions,
            selected: selected,
            self: self,
            display,
            ws: null,
            timer: null,
        }
    },

    methods: {
        checkSession(sid,session) {
            history.replaceState(null, null, '?select=' + sid);
            this.selected = session
        },
        refreshSession() {
            console.log('重加载消息  ')
            this.sessions = getSessions()
        }

    },
    mounted() {
        const t = this
        let scheme = "ws://"
        if (window.location.protocol === 'https:') {
            scheme = "wss://"
        }
        const ws = new WebSocket(scheme + window.location.host + "/api/ws");
        ws.onopen = function (evt) {
            console.log("event connected.");
        }
        ws.onclose = function (evt) {
            t.ws = null;
        }
        ws.onmessage = function (evt) {
            console.log("event received.");
            t.refreshSession()
        }
        ws.onerror = function (evt) {
            console.log("ERROR: " + evt.data);
        }
        this.ws = ws;
        this.timer = setInterval(this.refreshSession, 10*1000);
    },
    beforeDestroy() {
        this.ws.close()
        this.ws = null;
        this.timer = null;
    }
})
app.component('session', {
    props: ['session', 'self', 'active'],
    emits: ['change'],
    methods:{
        timeFMT(time){
            return timeFormat(time)
        },
        msgFMT(msg){
            return plainMessage(msg)
        },

    },
    template: `
      <div class="chat-item" :class="{active:active}" @click="$emit('change')">
        <div class="avatar">
          <img :src="session.avatar" alt="用户头像">
        </div>
        <div class="chat-info">
          <div class="chat-info-header">
            <div class="chat-name"><span v-if="session.type === 'group'">[群]</span>{{ session.name }}</div>
            <div class="chat-time">{{ timeFMT(session.last_message.created) }}</div>
          </div>
          <div class="chat-message">
            <template v-if="session.last_message.sender.id !== self.id ">{{session.last_message.sender.name}}:</template>
            {{ msgFMT(session.last_message.content) }}</div>
        </div>
      </div>
    `,
})

app.component('chat-header', {
    props: ['title'],
    emits: ['back'],
    setup(props) {

    },
    template: `
      <div class="chat-header">
        <div class="back-button" @click="$emit('back')">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M15 19L9 12L15 5" stroke="#333" stroke-width="2" stroke-linecap="round"
                  stroke-linejoin="round"></path>
          </svg>
        </div>
        <div class="current-chat-name">{{ title }}</div>
      </div>
    `
})

app.component('chat-message-content', {
    props: ['content'],
    setup(props) {
        const refer = computed(() => {
            if (props.content.refer_id !== "") {
                const msg = getMessage(props.content.refer_id)
                return {
                    name: msg.content.sender.name,
                    created: timeFormat(msg.content.created),
                    message: plainMessage(msg.content.content)
                }
            } else {
                return null
            }
        });

        return {
            refer
        }
    },
    methods: {
        atUserInfo(id) {
            return getUserInfo(id)
        }
    },
    template: `
      <div class="message-content">
        <div class="message-reference" v-if="refer != null">
          <div class="message-reference-sender"><span class="message-reference-name">{{ refer.name }}</span><span
              class="message-reference-time">{{ refer.created }}</span></div>
          <div class="message-reference-content">
            {{ refer.message }}
          </div>
        </div>
        <div class="message-main-content">
          <template v-for="item in content.content">
            <span v-if="item.type ==='text'">{{ item.data }}</span>
            <span v-if="item.type ==='at'" class="message-at">@{{ atUserInfo(item.data).name }}</span>
            <div class="message-image" v-if="item.type ==='image'">
              <img alt="" :src="item.data"/>
            </div>
          </template>
        </div>
      </div>
    `
})

app.component('chat-message', {
    props: [ 'content','self'],
    emits: ['refer'],
    methods:{
        timeFMT(time){
            return timeFormat(time)
        },
    },
    template: `
      <div style="position: relative" class="message" :class="content.sender.id ===self.id ?'sent':'received'">
        <div class="avatar"><img :src="content.sender.avatar" alt="用户头像"></div>
        <div class="message-group">
          <p class="message-group-name">{{ content.sender.alias?content.sender.alias:content.sender.name }}</p>
          <chat-message-content :content="content"></chat-message-content>
          <div class="message-time">{{ timeFMT(content.created) }}</div>
        </div>
        <div class="reference-icon" @click="$emit('refer')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M10 9V5l-7 7 7 7v-4.1c5 0 8.5 1.6 11 5.1-1-5-4-10-11-11z" fill="#888"/>
          </svg>
        </div>
      </div>
    `
})

app.component('chat-input-area', {
    props: ['refer', 'session'],
    emits: ['update:refer', 'send'],
    setup(props) {
        const message = ref("")
        const images = ref([])
        const ats = ref([])
        watch(
            () => props.session,
            () => {
                message.value = '';
                images.value = [];
            },
        );
        return {
            images: images,
            message: message,
            ats: ats,
        }
    },
    methods: {
        removeImage(index) {
            this.images.splice(index, 1); // 从数组中移除
        },
        paste(e) {
            if (e.clipboardData && e.clipboardData.items) {
                for (let i = 0; i < e.clipboardData.items.length; i++) {
                    const item = e.clipboardData.items[i];
                    // 检查项目是否为图片
                    if (item.type.indexOf('image') !== -1) {
                        const blob = item.getAsFile();
                        const reader = new FileReader();
                        // 读取图片为Base64格式
                        reader.onload = (event) => {
                            // 创建图片元素
                            const imageUrl = event.target.result;
                            if (!this.images.includes(imageUrl)) {
                                this.images.push(imageUrl)
                            }
                        };
                        reader.readAsDataURL(blob);
                        e.preventDefault();
                    }
                }
            }
        },
        renderAt(text, index) {
            const lastAtIndex = text.lastIndexOf('@', index - 1);
            if (lastAtIndex === -1 || (lastAtIndex > 0 && /\S/.test(text[lastAtIndex - 1]))) {
                this.ats = []
                return;
            }
            const query = text.substring(lastAtIndex + 1, index);
            this.ats = searchUser(this.session, query.toLowerCase());
        },
        atClick(text, user) {
            const index = this.$refs.input.selectionStart;
            const mentionText = `@${user.id} `;
            const lastAtIndex = text.lastIndexOf('@', index - 1);
            this.message = text.substring(0, lastAtIndex) + mentionText + text.substring(index);
            this.ats = []
            this.$refs.input.focus()
        },
        cleanupMsg(e, text) {
            if (e.key === 'Backspace') {
                const index = this.$refs.input.selectionStart;
                const lastAtIndex = text.lastIndexOf('@', index - 1);
                if (lastAtIndex >= 0 && !(text.substring(lastAtIndex, text.length).includes(' '))) {
                    this.ats = []
                    this.message = text.substring(0, lastAtIndex);
                }
            }
        },

        sendMessage() {
            if (this.message === '' && this.images.length === 0) {
                return
            }
            let pushData = {
                session_id: this.session.id,
                user_id: selfInfo().id,
                refer_id: '',
                content: parseMessage(this.message),
            };
            if (this.refer != null) {
                pushData.refer_id = this.refer.id
            }
            for (let image of this.images) {
                pushData.content.push({
                    type: 'image',
                    data: image,
                })
            }
            postObj('/api/send', pushData)
            this.message = ''
            this.images = []
            this.$emit('send', null)
            this.$emit('update:refer', null)
            this.$refs.input.focus()
        },
        plainMessage(msg) {
            return plainMessage(msg)
        }
    },
    template: `
      <div class="chat-input-area">
        <div class="referenced-message" v-if="refer != null">
          <div class="referenced-message-header">
            <div class="referenced-message-sender">{{ refer.name }}</div>
            <div class="referenced-message-time">{{ refer.created }}</div>
            <div class="remove-reference" @click="$emit('update:refer',null)">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path
                    d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"
                    fill="#888"/>
              </svg>
            </div>
          </div>
          <div class="referenced-message-content">{{ plainMessage(refer.content) }}</div>
        </div>
        <div class="attachments-container">
          <div
              v-for="(url, index) in images"
              :key="index"
              class="attachment-preview"
              :data-image-url="url"
          >
            <img :src="url" alt="预览图">
            <div
                class="remove-attachment"
                @click="removeImage(index)"
            >×
            </div>
          </div>
        </div>
        <div class="chat-input-container">
          <label for="chatInput">
          </label><textarea ref="input" class="chat-input" placeholder="输入消息或粘贴图片..." v-model="message"
                            @input="e=>renderAt(message,e.target.selectionStart)"
                            @keyup="e=>cleanupMsg(e,message)"
                            @keydown.esc="ats=[]"
                            @keydown.enter="e=>{
                               if (!e.shiftKey ) {
                                    e.preventDefault();
                                   sendMessage();
                               }
                            }"
                            @paste="e=>paste(e)"></textarea>
          <button class="send-button" id="sendButton" @click="sendMessage()">
            <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" fill="none" viewBox="0 0 24 24">
              <path fill="currentColor"
                    d="m3.543 8.883 7.042-7.047a2 2 0 0 1 2.828 0l7.043 7.046a1 1 0 0 1 0 1.415l-.701.701a1 1 0 0 1-1.414 0L13.3 5.956v15.792a1 1 0 0 1-1 1h-.99a1 1 0 0 1-1-1V6.342l-4.654 4.656a1 1 0 0 1-1.414 0l-.7-.7a1 1 0 0 1 0-1.415"></path>
            </svg>
          </button>
        </div>
        <!-- @用户下拉列表 -->
        <div class="mention-dropdown" v-if="ats.length > 0">
          <div class="mention-item" v-for="user in ats" @click="atClick(message,user)">
            <div class="avatar">
              <img :src="user.avatar" :alt="user.groupName?user.groupName:user.name">
            </div>
            <div class="name">{{ user.groupName ? user.groupName : user.name }}</div>
            <div class="id">(<span class="sid">{{ user.id }}</span>)</div>
          </div>
        </div>
      </div>
    `
})

app.component('chat', {
    props: ['sessions', 'session', 'self'],
    emits: ['back', 'on-update'],
    setup(props) {
        // 定义计算属性，自动响应 session 的变化
        const messages = ref({})
        const title = ref("")
        const refer = ref(null)
        const refresh = function (props) {
            refer.value = null;
            if (props.session.type === "group") {
                const g = getGroupInfo(props.session.target);
                title.value = g.name + `  (${Object.keys(g.users).length})`;
            } else {
                title.value = props.session.name
            }
            messages.value = getMessages(props.session.id)
            // this.scrollToBottom();
        }
        refresh(props)
        watch(
            () => props.session,
            () => {
                refresh(props)
            },
        );
        watch(
            () => props.sessions,
            () => {
                refresh(props)
            },
        );
        return {
            title,
            messages,
            refer,
        };
    },
    watch: {
        // 当列表数据有变动时，滚动到底部
        messages() {
            this.$nextTick(() => {
                this.scrollToBottom();
            });
        }
    },
    mounted() {
        // 页面初次加载时就滚动到底部
        this.scrollToBottom();
    },
    methods: {
        checkReference(messageId) {
            const msg = getMessage(messageId)
            this.refer = {
                id: messageId,
                name: msg.content.sender.name,
                avatar: msg.content.sender.avatar,
                content: msg.content.content,
                created: timeFormat(msg.content.created),
            }
        },
        scrollToBottom() {
            const container = this.$refs.scrollContainer;
            if (container) {
                container.scrollTop = container.scrollHeight;
            }
        }
    },
    template: `
      <div class="chat-container" :class="session.type === 'group'?'chat-group-session':'chat-private-session'">
        <chat-header @back="$emit('back')" :title="title"></chat-header>
        <div class="chat-messages" ref="scrollContainer">
          <chat-message v-for="message in messages.content"
                        :content="message"
                        :self="self"
                        @refer="checkReference(message.id)"
          ></chat-message>
        </div>
        <chat-input-area :session="session" v-model:refer="refer" @send="$emit('on-update')"></chat-input-area>
      </div>
    `
})

app.mount('#app')