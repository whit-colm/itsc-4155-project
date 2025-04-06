-- Actually removing a comment from the database can cause god-knows-
-- what issues. So instead on delete we remove key values and manually
-- 'cascade' votes.
CREATE OR REPLACE FUNCTION comment_faux_delete()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE comments 
    SET deleted = true, poster_id = NULL, body = NULL, rating = NULL
    WHERE id = OLD.id;

    -- Delete 
    DELETE FROM votes
    WHERE comment_id = OLD.id;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- New votes
CREATE OR REPLACE FUNCTION update_vote_total_insert()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE comments
    SET vote_total = vote_total + NEW.vote
    WHERE id = NEW.comment_id AND deleted = false;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- When a vote changes
CREATE OR REPLACE FUNCTION update_vote_total_update()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE comments 
    SET vote_total = vote_total - OLD.vote + NEW.vote 
    WHERE id = NEW.comment_id AND deleted = false;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- When a vote is removed
CREATE OR REPLACE FUNCTION update_vote_total_delete()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE comments 
    SET vote_total = vote_total - OLD.vote 
    WHERE id = OLD.comment_id AND deleted = false;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;



CREATE TRIGGER t_comments_delete
BEFORE DELETE ON comments
FOR EACH ROW EXECUTE FUNCTION comment_faux_delete();

CREATE TRIGGER t_votes_insert
AFTER INSERT ON votes
FOR EACH ROW EXECUTE FUNCTION update_vote_total_insert();

CREATE TRIGGER t_votes_update
AFTER UPDATE ON votes
FOR EACH ROW EXECUTE FUNCTION update_vote_total_update();

CREATE TRIGGER t_votes_delete
AFTER DELETE ON votes
FOR EACH ROW EXECUTE FUNCTION update_vote_total_delete();