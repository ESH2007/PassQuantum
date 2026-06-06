import { LangProvider } from './LangContext'
import Navbar from './components/Navbar'
import Hero from './components/Hero'
import Features from './components/Features'
import Architecture from './components/Architecture'
import Footer from './components/Footer'

function App() {
  return (
    <LangProvider>
      <div className="min-h-screen bg-bg text-slate-300 overflow-x-hidden">
        <Navbar />
        <main>
          <Hero />
          <Features />
          <Architecture />
        </main>
        <Footer />
      </div>
    </LangProvider>
  )
}

export default App
