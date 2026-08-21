package com.kingfisher.modules.agent.mapper;

import com.kingfisher.modules.agent.domain.AgentMessage;
import com.kingfisher.modules.agent.domain.Conversation;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

/**
 * Agent Mapper，映射 SQLite agent_conversations / agent_messages 表。
 */
@Mapper
public interface AgentMapper {

    List<Conversation> listConversations(@Param("userId") Long userId);

    Conversation findConversationById(@Param("id") Long id, @Param("userId") Long userId);

    void insertConversation(Conversation conversation);

    void deleteConversation(@Param("id") Long id, @Param("userId") Long userId);

    void deleteMessagesByConversation(@Param("conversationId") Long conversationId);

    List<AgentMessage> listMessages(@Param("conversationId") Long conversationId, @Param("userId") Long userId);

    void insertMessage(AgentMessage message);

    void updateConversationTitle(@Param("id") Long id, @Param("title") String title);
}
