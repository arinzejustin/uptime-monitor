import { handle } from '@hono/node-server/vercel'
import app from '../src/index.ts'

export default handle(app);