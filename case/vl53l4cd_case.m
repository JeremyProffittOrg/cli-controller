% Adafruit VL53L4CD (product 5396) ToF case — MATLAB drawing
% Units: millimetres. Run in MATLAB: vl53l4cd_case
% Open back, 12 mm deep, 39 mm long sides with r=20 half-circle wings,
% rounded-square window, 3 mm bosses with 2 mm x 4 mm holes (not through front).

function vl53l4cd_case
    P = case_params();
    fig = figure('Color','w','Name','VL53L4CD ToF case','Position',[80 60 1480 920]);
    tiledlayout(fig,2,3,'Padding','compact','TileSpacing','compact');

    nexttile; draw_front(P);
    title('Front  (X-Z)  sensor face'); xlabel('X mm'); ylabel('Z mm');
    axis equal tight; grid on; box on;

    nexttile; draw_top(P);
    title('Top  (X-Y)'); xlabel('X mm'); ylabel('Y mm');
    axis equal tight; grid on; box on;

    nexttile; draw_right(P);
    title('Right  (Y-Z)'); xlabel('Y mm'); ylabel('Z mm');
    axis equal tight; grid on; box on;

    nexttile; draw_back(P);
    title('Back  (X-Z)  open'); xlabel('X mm'); ylabel('Z mm');
    axis equal tight; grid on; box on;

    nexttile([1 2]);
    draw_iso(P);
    title(sprintf('Isometric  |  %.0f x %.0f x %.2f mm + r=20 wings', P.outer_x, P.depth, P.outer_z));
    xlabel('X mm'); ylabel('Y mm'); zlabel('Z mm');
    axis equal vis3d; grid on; box on;
    view(38, 22); camlight('headlight'); lighting gouraud;

    sgtitle(fig, 'Adafruit VL53L4CD (5396) time-of-flight case  (mm)');
    out = fullfile(fileparts(mfilename('fullpath')), 'vl53l4cd_case_matlab.png');
    exportgraphics(fig, out, 'Resolution', 150);
    fprintf('Wrote %s\n', out);
end

function P = case_params()
    P.pcb_x = 25.40; P.pcb_z = 17.78; P.pcb_t = 1.60;
    P.hole_span_x = 20.32; P.hole_span_z = 12.70;
    P.pcb_hole_d = 2.50;
    P.sensor_x = 4.40; P.sensor_z = 2.40;
    P.wall = 3.00; P.front_t = 3.00; P.fit = 0.50;
    P.outer_x = 39.00; P.depth = 12.00;
    P.cut = 3.00; P.cut_r = 0.60; P.cut_base_r = 1.50; P.cut_y0 = 0.00;
    P.window_s = 10.00; P.window_r = 2.00;
    P.boss_h = 3.00; P.boss_od = 5.00; P.boss_hole = 2.00; P.boss_hole_depth = 4.00;
    P.wing_t = 4.00; P.wing_r = 12.00; P.wing_hole_d = 4.00; P.wing_hole_y = 6.00;
    P.inner_x = P.outer_x - 2*P.wall;
    P.inner_z = P.pcb_z + 2*P.fit;
    P.outer_z = P.inner_z + 2*P.wall;
end

function draw_front(P)
    hold on;
    ox = P.outer_x/2; oz = P.outer_z/2;
    rectangle('Position',[-ox -oz P.outer_x P.outer_z],'EdgeColor','k','LineWidth',1.4);
    rectangle('Position',[-P.window_s/2 -P.window_s/2 P.window_s P.window_s], ...
        'Curvature',[P.window_r/(P.window_s/2) P.window_r/(P.window_s/2)], ...
        'EdgeColor',[0.85 0.33 0.1],'LineWidth',1.8);
    for sx = [-1 1]
        for sz = [-1 1]
            viscircles([sx*P.hole_span_x/2, sz*P.hole_span_z/2], P.boss_hole/2, ...
                'Color',[0 0.447 0.741], 'LineWidth',1.2);
        end
        t = linspace(0, pi, 40);
        plot(P.wing_r*cos(t), sx*(oz + P.wing_r*sin(t)), 'Color',[0.49 0.18 0.56], 'LineWidth',1.6);
        viscircles([0, sx*(oz+P.wing_hole_y)], P.wing_hole_d/2, 'Color',[0.49 0.18 0.56], 'LineWidth',1.4);
    end
    dimh(0, -ox, ox, oz+P.wing_r+6, sprintf('%.0f', P.outer_x));
    dimv(-ox-8, -oz, oz, sprintf('%.2f', P.outer_z));
    hold off;
end

function draw_top(P)
    hold on;
    ox = P.outer_x/2;
    rectangle('Position',[-ox 0 P.outer_x P.depth],'EdgeColor','k','LineWidth',1.4);
    plot([-ox ox],[P.front_t P.front_t],'Color',[0.2 0.2 0.7],'LineStyle','--');
    for s = [-1 1]
        rectangle('Position',[s*(ox-P.wall), P.cut_y0, s*P.wall, P.cut], ...
            'EdgeColor',[0.47 0.67 0.19],'LineWidth',1.4);
    end
    dimh(0, -ox, ox, -6, sprintf('%.0f', P.outer_x));
    dimv(ox+6, 0, P.depth, sprintf('%.0f', P.depth));
    hold off;
end

function draw_right(P)
    hold on;
    oz = P.outer_z/2;
    rectangle('Position',[0 -oz P.depth P.outer_z],'EdgeColor','k','LineWidth',1.4);
    rectangle('Position',[0 oz P.wing_t P.wing_r],'EdgeColor',[0.49 0.18 0.56],'LineWidth',1.3);
    rectangle('Position',[0 -oz-P.wing_r P.wing_t P.wing_r],'EdgeColor',[0.49 0.18 0.56],'LineWidth',1.3);
    rectangle('Position',[P.cut_y0, -P.cut/2, P.cut, P.cut], ...
        'Curvature',[P.cut_r/(P.cut/2) P.cut_r/(P.cut/2)], ...
        'EdgeColor',[0.47 0.67 0.19],'LineWidth',1.6);
    dimh(oz+P.wing_r+6, 0, P.depth, sprintf('%.0f', P.depth));
    hold off;
end

function draw_back(P)
    hold on;
    ox = P.outer_x/2; oz = P.outer_z/2;
    rectangle('Position',[-ox -oz P.outer_x P.outer_z],'EdgeColor','k','LineWidth',1.4);
    rectangle('Position',[-P.window_s/2 -P.window_s/2 P.window_s P.window_s], ...
        'Curvature',[P.window_r/(P.window_s/2) P.window_r/(P.window_s/2)], ...
        'EdgeColor',[0.85 0.33 0.1],'LineWidth',1.5);
    hold off;
end

function draw_iso(P)
    hold on;
    c = [0 0.447 0.741];
    x = [-P.outer_x/2 P.outer_x/2]; y = [0 P.depth]; z = [-P.outer_z/2 P.outer_z/2];
    patch('XData',[x(1) x(2) x(2) x(1)],'YData',[y(1) y(1) y(1) y(1)],'ZData',[z(1) z(1) z(2) z(2)], ...
        'FaceColor',c,'FaceAlpha',0.55,'EdgeColor','k');
    patch('XData',[x(1) x(1) x(1) x(1)],'YData',[y(1) y(2) y(2) y(1)],'ZData',[z(1) z(1) z(2) z(2)], ...
        'FaceColor',c,'FaceAlpha',0.35,'EdgeColor','k');
    patch('XData',[x(2) x(2) x(2) x(2)],'YData',[y(1) y(2) y(2) y(1)],'ZData',[z(1) z(1) z(2) z(2)], ...
        'FaceColor',c,'FaceAlpha',0.50,'EdgeColor','k');
    patch('XData',[x(1) x(2) x(2) x(1)],'YData',[y(1) y(1) y(2) y(2)],'ZData',[z(2) z(2) z(2) z(2)], ...
        'FaceColor',[0.3 0.3 0.35],'FaceAlpha',0.55,'EdgeColor','k');
    patch('XData',[x(1) x(2) x(2) x(1)],'YData',[y(1) y(1) y(2) y(2)],'ZData',[z(1) z(1) z(1) z(1)], ...
        'FaceColor',c,'FaceAlpha',0.25,'EdgeColor','k');
    hold off;
end

function dimh(~, x1, x2, y, label)
    plot([x1 x2],[y y],'k-');
    plot(x1,y,'k<','MarkerFaceColor','k','MarkerSize',4);
    plot(x2,y,'k>','MarkerFaceColor','k','MarkerSize',4);
    text((x1+x2)/2, y+0.5, label, 'HorizontalAlignment','center', 'FontSize',8, 'BackgroundColor','w');
end

function dimv(x, y1, y2, label)
    plot([x x],[y1 y2],'k-');
    plot(x,y1,'kv','MarkerFaceColor','k','MarkerSize',4);
    plot(x,y2,'k^','MarkerFaceColor','k','MarkerSize',4);
    text(x+0.5, (y1+y2)/2, label, 'HorizontalAlignment','left', 'FontSize',8, 'BackgroundColor','w');
end
