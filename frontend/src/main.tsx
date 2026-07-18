import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'
import { initializeTheme } from './shared/theme/theme'

;(window as Window & { __ANT_APP_BOOTED__?: boolean }).__ANT_APP_BOOTED__ = true
initializeTheme()

ReactDOM.createRoot(document.getElementById('root')!).render(<App />)

