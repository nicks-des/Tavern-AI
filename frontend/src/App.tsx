import { useEffect } from 'react'
import { useStore } from './store'
import { Sidebar } from './components/Sidebar'
import { WelcomeScreen } from './components/WelcomeScreen'
import { ChatArea } from './components/ChatArea'
import { CharacterManager } from './components/CharacterManager'
import { RoomDetail } from './components/RoomDetail'

export default function App() {
  const { activeSessionId, activeRoomId, currentView, setCurrentView, loadCharacters } = useStore()

  useEffect(() => {
    loadCharacters()
  }, [loadCharacters])

  return (
    <div className="flex h-screen w-screen">
      <Sidebar />
      <main className="flex-1 flex flex-col min-w-0">
        {currentView === 'manage' ? (
          <CharacterManager onClose={() => setCurrentView('chat')} />
        ) : activeRoomId ? (
          <RoomDetail />
        ) : activeSessionId ? (
          <ChatArea />
        ) : (
          <WelcomeScreen />
        )}
      </main>
    </div>
  )
}
