const {createApp, ref, computed, watch} = Vue
let app = createApp({
    setup() {
        const sessions = ref(getSessions())
        const self = ref(selfInfo())
        const selected = ref(null)
        const display = ref(false)
        const params = new URLSearchParams(window.location.search);
        let s = params.get("select")
        if (s == null || s === "") {
            selected.value = getSessions()[0]
        } else {
            const s2 = getSession(s);
            if (s2 == null) {
                selected.value = getSessions()[0]
            } else {
                selected.value = s2
            }
        }
        return {
            sessions: sessions,
            selected: selected,
            self: self,
            display
        }
    },

    methods: {
        checkSession(session) {
            history.replaceState(null, null, '?select=' + session.id);
            this.selected = session
        }
    }
})
app.component('session', {
    props: ['session', 'self', 'active'],
    emits: ['change'],
    setup(props) {
        const meta = ref({
            title: "",
            image: "",
            isGroup: false,
            message: "",
            lastUpdate: "",
        })
        if (props.session.type === "group") {
            meta.value.isGroup = true
            const groupInfo = getGroupInfo(props.session.target);
            const lastMsg = getLastMessage(props.session.id);
            meta.value.title = groupInfo.name
            meta.value.image = groupInfo.image
            meta.value.lastUpdate = timeFormat(lastMsg.time)
            if (lastMsg.sender !== self.id) {
                meta.value.message = getUserInfo(lastMsg.sender).name + ":" + plainMessage(lastMsg.content.message)
            } else {
                meta.value.message = plainMessage(lastMsg.content.message)
            }
        } else {
            const userInfo = getUserInfo(props.session.target);
            const lastMsg = getLastMessage(props.session.id);
            meta.value.title = userInfo.name
            meta.value.image = userInfo.image
            meta.value.message = plainMessage(lastMsg.content.message)
            meta.value.lastUpdate = timeFormat(lastMsg.time)
        }
        return {
            meta: meta
        }
    },
    template: `
      <div class="chat-item" :class="{active:active}" @click="$emit('change')">
        <div class="avatar">
          <img :src="meta.image" alt="用户头像">
        </div>
        <div class="chat-info">
          <div class="chat-info-header">
            <div class="chat-name">{{ meta.title }}<span v-if="meta.isGroup">(群组)</span></div>
            <div class="chat-time">{{ meta.lastUpdate }}</div>
          </div>
          <div class="chat-message">{{ meta.message }}</div>
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
            if (props.content.ref != null) {
                const msg = getMessage(props.content.ref.sid, props.content.ref.mid)
                return {
                    name: getUserInfo(msg.sender).name,
                    message: plainMessage(msg.content.message),
                    lastUpdate: timeFormat(msg.time)
                }
            } else {
                return null
            }
        })
        return {
            refer
        }
    },
    template: `
      <div class="message-content">
        <div class="message-reference" v-if="content.ref != null && refer != null">
          <div class="message-reference-sender"><span class="message-reference-name">{{ refer.name }}</span><span
              class="message-reference-time">{{ refer.lastUpdate }}</span></div>
          <div class="message-reference-content">
            {{ refer.message }}
          </div>
        </div>
        <div class="message-main-content">
          <template v-for="item in content.message">
            <span v-if="item.type ==='text'">{{ item.data }}</span>
            <span v-if="item.type ==='at'" class="message-at">@{{ item.id }}</span>
            <div class="message-image" v-if="item.type ==='image'">
              <img alt="" :src="item.data"/>
            </div>
          </template>
        </div>
      </div>
    `
})

app.component('chat-message', {
    props: ['name', 'image', 'content', 'isSelf', 'updateTime'],
    emits: ['refer'],
    template: `
      <div style="position: relative" class="message" :class="isSelf?'sent':'received'">
        <div class="avatar"><img :src="image" alt="用户头像"></div>
        <div class="message-group">
          <p class="message-group-name">{{name}}</p>
          <chat-message-content :content="content"></chat-message-content>
          <div class="message-time">{{ updateTime }}</div>
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
    emits: ['update:refer'],
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
            this.ats = searchUser(this.session.id, query.toLowerCase());
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
            if (this.message === '') {
                return
            }
            let msg = {
                refer: null,
                message: parseMessage(this.message),
            };
            if (this.refer != null) {
                msg.refer = {
                    sid: this.refer.sid,
                    mid: this.refer.mid,
                }
            }
            for (let image of this.images) {
                msg.message.push({
                    type: 'image',
                    file: image,
                })
            }

            console.log(this.message, JSON.stringify(msg));
            this.message = ''
            this.images = []
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
            <div class="referenced-message-time">{{ refer.lastUpdate }}</div>
            <div class="remove-reference" @click="$emit('update:refer',null)">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path
                    d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"
                    fill="#888"/>
              </svg>
            </div>
          </div>
          <div class="referenced-message-content">{{ plainMessage(refer.content.message) }}</div>
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
          </label><textarea ref="input" class="chat-input" placeholder="输入消息..." v-model="message"
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
              <img :src="user.image" :alt="user.name">
            </div>
            <div class="name">{{ user.name }}</div>
            <div class="id">(<span class="sid">{{ user.id }}</span>)</div>
          </div>
        </div>
      </div>
    `
})

app.component('chat', {
    props: ['session', 'self'],
    emits: ['back'],
    setup(props) {
        // 定义计算属性，自动响应 session 的变化
        const messages = computed(() => {
            return getMessages(props.session.id);
        })
        const title = computed(() => {
            if (props.session.type === "group") {
                const g = getGroupInfo(props.session.target);
                return g.name + `  (${g.users.length})`;
            } else {
                const u = getUserInfo(props.session.target);
                return u.name;
            }
        });
        const refer = ref(null)
        watch(
            () => props.session,
            () => {
                refer.value = null;
            },
        );
        return {
            title,
            messages,
            refer
        };
    },
    methods: {
        checkReference(sessionID, messageId) {
            const msg = getMessage(sessionID, messageId)
            let userInfo = getUserInfo(msg.sender);
            this.refer = {
                sid: sessionID,
                mid: messageId,
                name: userInfo.name,
                image: userInfo.image,
                content: msg.content,
                lastUpdate: timeFormat(msg.time),
            }
        }
    },
    template: `
      <div class="chat-container" :class="session.type === 'group'?'chat-group-session':'chat-private-session'">
        <chat-header @back="$emit('back')" :title="title"></chat-header>
        <div class="chat-messages">
          <chat-message v-for="message in messages"
                        :name="message.name"
                        :content="message.content"
                        :image="message.image"
                        :isSelf="message.isSelf"
                        :update-time="message.lastUpdate"
                        @refer="checkReference(session.id,message.id)"
          ></chat-message>
        </div>
        <chat-input-area :session="session" v-model:refer="refer"></chat-input-area>
      </div>
    `
})

app.mount('#app')