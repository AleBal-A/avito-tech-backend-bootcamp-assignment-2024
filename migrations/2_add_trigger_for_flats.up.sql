CREATE OR REPLACE FUNCTION update_last_flat_added()
    RETURNS TRIGGER AS $$
BEGIN
    UPDATE houses
    SET last_flat_added = CURRENT_TIMESTAMP
    WHERE id = NEW.house_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER after_flat_insert
    AFTER INSERT ON flats
    FOR EACH ROW
EXECUTE FUNCTION update_last_flat_added();
